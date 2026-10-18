package migration

import (
	"strings"
	"testing"
)

// buildFor renders a table without touching DB_DRIVER, so these cases can run
// in any order alongside tests that do set it.
func buildFor(t *testing.T, driver string, fn func(*Table)) string {
	t.Helper()
	return normalizeSchema(CreateFor(driver, "posts", fn).Build())
}

// A unique index is not a unique constraint, and only one of the two was
// sayable. Postgres keeps them as different catalog objects, so a migration
// that declared an index could not be restated with UniqueKey — pg_dump
// renders the constraint as ALTER TABLE ... ADD CONSTRAINT and the index as
// CREATE UNIQUE INDEX, and a schema being *matched* rather than merely
// created has to produce the right one.
func TestAUniqueIndexCanBeDeclared(t *testing.T) {
	for driver, want := range map[string]string{
		DriverPostgres: "CREATE UNIQUE INDEX posts_slug_idx ON posts (slug);",
		DriverSQLite:   "CREATE UNIQUE INDEX posts_slug_idx ON posts (slug);",
		// MySQL takes it inside the table. The others cannot: inline INDEX is
		// a MySQL extension.
		DriverMySQL: "UNIQUE INDEX posts_slug_idx (slug)",
	} {
		t.Run(driver, func(t *testing.T) {
			sql := buildFor(t, driver, func(tb *Table) {
				tb.BigIncrements("id").Primary()
				tb.String("slug", 120)
				tb.Index("slug").Unique().Name("posts_slug_idx")
			})
			if !strings.Contains(sql, normalizeSchema(want)) {
				t.Fatalf("expected %q in:\n%s", want, sql)
			}
		})
	}
}

// And an index without Unique stays an ordinary one, so the default is not
// quietly promoted.
func TestAnIndexIsNotUniqueUnlessAsked(t *testing.T) {
	sql := buildFor(t, DriverPostgres, func(tb *Table) {
		tb.String("slug", 120)
		tb.Index("slug").Name("posts_slug_idx")
	})
	if strings.Contains(sql, normalizeSchema("UNIQUE")) {
		t.Fatalf("a plain index came out unique:\n%s", sql)
	}
}

// The two spellings produce different SQL, which is the whole reason both
// exist. A test rather than a comment, because "they are the same thing on
// MySQL" is true and is what makes the difference easy to lose.
func TestAUniqueIndexAndAUniqueKeyAreNotTheSameStatement(t *testing.T) {
	asIndex := buildFor(t, DriverPostgres, func(tb *Table) {
		tb.String("slug", 120)
		tb.Index("slug").Unique().Name("posts_slug_key")
	})
	asConstraint := buildFor(t, DriverPostgres, func(tb *Table) {
		tb.String("slug", 120)
		tb.UniqueKey("slug").Name("posts_slug_key")
	})
	if asIndex == asConstraint {
		t.Fatalf("Index().Unique() and UniqueKey() produced identical SQL:\n%s", asIndex)
	}
	if !strings.Contains(asIndex, normalizeSchema("CREATE UNIQUE INDEX")) {
		t.Errorf("the index form is not an index:\n%s", asIndex)
	}
	if !strings.Contains(asConstraint, normalizeSchema("CONSTRAINT posts_slug_key UNIQUE")) {
		t.Errorf("the constraint form is not a constraint:\n%s", asConstraint)
	}
}

// Postgres, MySQL 8.0 and SQLite 3.8.3 all honour a descending key, and the
// builder had no word for one — so every "newest first" index in an
// application was raw SQL written for one vendor.
func TestAnIndexColumnCanBeDescending(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			sql := buildFor(t, driver, func(tb *Table) {
				tb.UnsignedBigInt("user_id")
				tb.TimestampTz("created_at", 0)
				tb.Index("user_id", "created_at").
					Desc("created_at").
					Name("posts_author_recent_idx")
			})
			if !strings.Contains(sql, normalizeSchema("(user_id, created_at DESC)")) {
				t.Fatalf("expected a descending key in:\n%s", sql)
			}
			// Only the column asked for, and the key order is the Index order
			// rather than the Desc order.
			if strings.Contains(sql, normalizeSchema("user_id DESC")) {
				t.Fatalf("a column nobody named came out descending:\n%s", sql)
			}
		})
	}
}

// Desc naming a column the index does not carry is the quiet failure this
// method could have shipped with: an index silently sorted the way it always
// was, and a query plan nobody can account for.
func TestOrderingAColumnTheIndexDoesNotCarryIsRefused(t *testing.T) {
	schema := CreateFor(DriverPostgres, "posts", func(tb *Table) {
		tb.UnsignedBigInt("user_id")
		tb.TimestampTz("created_at", 0)
		tb.Index("user_id").Desc("created_at")
	})
	err := schema.Validate()
	if err == nil {
		t.Fatal("expected ordering an unindexed column to be refused")
	}
	for _, want := range []string{"created_at", "does not index it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q: %v", want, err)
		}
	}
}

// A composite foreign key: one constraint over a pair, which two
// single-column keys do not express — neither of them requires the pair to
// exist together, which is the invariant the pivot is there to hold.
func TestAForeignKeyCanReferenceSeveralColumns(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			sql := buildFor(t, driver, func(tb *Table) {
				tb.UnsignedBigInt("user_id")
				tb.UnsignedBigInt("skill_id")
				tb.Foreign("user_id", "skill_id").
					References("user_id", "skill_id").
					On("user_skills").
					OnDelete("CASCADE")
			})
			want := "FOREIGN KEY (user_id, skill_id) REFERENCES user_skills(user_id, skill_id) ON DELETE CASCADE"
			if !strings.Contains(sql, normalizeSchema(want)) {
				t.Fatalf("expected %q in:\n%s", want, sql)
			}
		})
	}
}

// The single-column call is unchanged, which is what makes this additive: the
// variadic signature accepts every existing References("id").
func TestASingleColumnForeignKeyIsUnchanged(t *testing.T) {
	sql := buildFor(t, DriverPostgres, func(tb *Table) {
		tb.UnsignedBigInt("user_id")
		tb.Foreign("user_id").References("id").On("users").OnDelete("CASCADE")
	})
	want := "FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE"
	if !strings.Contains(sql, normalizeSchema(want)) {
		t.Fatalf("expected %q in:\n%s", want, sql)
	}
}

// A key carrying two columns and referencing one is not a narrower
// constraint, it is a syntax error — reported here rather than by the driver,
// where the message names neither method.
func TestAMismatchedCompositeForeignKeyIsRefused(t *testing.T) {
	schema := CreateFor(DriverPostgres, "posts", func(tb *Table) {
		tb.UnsignedBigInt("user_id")
		tb.UnsignedBigInt("skill_id")
		tb.Foreign("user_id", "skill_id").References("id").On("user_skills")
	})
	err := schema.Validate()
	if err == nil {
		t.Fatal("expected a mismatched composite foreign key to be refused")
	}
	if !strings.Contains(err.Error(), "as many columns as it carries") {
		t.Errorf("unexpected error: %v", err)
	}
}

// One declaration was producing three different schemas: Postgres got a
// CHECK, MySQL an UNSIGNED column type, SQLite nothing. For a surrogate key
// the portable answer is the dialect's own auto-increment type and no extra
// constraint — a sequence starting at 1 and ascending cannot violate it, and
// the CHECK made a generated table impossible to match with a hand-written
// one.
func TestAnAutoIncrementKeyCarriesNoCheckOnPostgres(t *testing.T) {
	for _, column := range []struct {
		name string
		add  func(*Table)
	}{
		{"BigIncrements", func(tb *Table) { tb.BigIncrements("id").Primary() }},
		{"Increments", func(tb *Table) { tb.Increments("id").Primary() }},
	} {
		t.Run(column.name, func(t *testing.T) {
			sql := buildFor(t, DriverPostgres, column.add)
			if strings.Contains(sql, normalizeSchema("CHECK")) {
				t.Fatalf("an auto-increment key still carries a CHECK:\n%s", sql)
			}
		})
	}
}

// MySQL keeps its UNSIGNED, because there the unsignedness is in the type
// rather than in a constraint. Dropping the CHECK must not have reached it.
func TestAnAutoIncrementKeyStaysUnsignedOnMySQL(t *testing.T) {
	sql := buildFor(t, DriverMySQL, func(tb *Table) {
		tb.BigIncrements("id").Primary()
	})
	if !strings.Contains(sql, normalizeSchema("BIGINT UNSIGNED")) {
		t.Fatalf("MySQL lost its unsigned auto-increment:\n%s", sql)
	}
}

// An explicit unsigned column emits no CHECK on Postgres either — and never
// did, which is the part worth pinning.
//
// Column.Unsigned sets the flag after NewDataType has already run
// AddSuffixes, so the suffix is decided before the column is marked. The same
// ordering bug Enum has. MySQL is unaffected, because its UNSIGNED is read
// when the type is rendered rather than baked into a suffix.
//
// Asserted rather than left implicit, so nobody reads the change above and
// concludes the CHECK moved somewhere: it was already unreachable here.
// Fixing it is a separate decision, because a counter wants >= 0 and the
// suffix says > 0, and either answer adds a constraint to tables already
// built.
func TestAnExplicitUnsignedColumnNeverCheckedOnPostgres(t *testing.T) {
	sql := buildFor(t, DriverPostgres, func(tb *Table) {
		tb.BigIncrements("id").Primary()
		tb.UnsignedInt("view_count")
	})
	if strings.Contains(sql, normalizeSchema("CHECK")) {
		t.Fatalf("unexpected CHECK — Column.Unsigned has started taking effect,"+
			" which is a behaviour change to decide on deliberately:\n%s", sql)
	}
	// MySQL does honour it, which is why the two dialects disagree.
	mysql := buildFor(t, DriverMySQL, func(tb *Table) {
		tb.UnsignedInt("view_count")
	})
	if !strings.Contains(mysql, normalizeSchema("INT UNSIGNED")) {
		t.Fatalf("MySQL stopped honouring an explicit unsigned column:\n%s", mysql)
	}
}

// The README's "Indexes, keys and constraints" example, kept here so it is
// compiled and executed rather than written from memory.
//
// The logging docs in this monorepo were fiction for a whole release —
// every sample called functions that did not exist — because nothing ever
// ran them. A doc example that builds on three dialects cannot drift from
// the API without this test going red.
func TestTheReadmeExampleBuilds(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			schema := CreateFor(driver, "endorsements", func(t *Table) {
				t.UnsignedBigInt("endorser_id")
				t.UnsignedBigInt("user_id")
				t.UnsignedBigInt("skill_id")
				t.String("note", 140).Default("")
				t.TimestampTz("created_at", 0).Default(Expr("CURRENT_TIMESTAMP"))

				t.Check("endorsements_not_self", "endorser_id <> user_id")

				t.Foreign("user_id", "skill_id").
					References("user_id", "skill_id").
					On("user_skills").
					OnDelete("CASCADE")

				t.Index("user_id", "created_at").Name("endorsements_recent_idx")

				t.Index("user_id", "created_at").
					Desc("created_at").
					Name("endorsements_feed_idx")

				t.Index("skill_id", "user_id").Unique().Name("endorsements_pair_key")

				t.UniqueKey("user_id", "note").Name("endorsements_one_note")
			})

			if err := schema.Validate(); err != nil {
				t.Fatalf("the README example does not validate: %v", err)
			}
			sql := normalizeSchema(schema.Build())

			// Each claim the prose makes about this example, asserted.
			for _, want := range []string{
				"CONSTRAINT endorsements_not_self CHECK (endorser_id <> user_id)",
				"FOREIGN KEY (user_id, skill_id) REFERENCES user_skills(user_id, skill_id)",
				"CONSTRAINT endorsements_one_note UNIQUE (user_id, note)",
				"(user_id, created_at DESC)",
			} {
				if !strings.Contains(sql, normalizeSchema(want)) {
					t.Errorf("expected %q in:\n%s", want, sql)
				}
			}

			// And the dialect split the prose leads with.
			if driver == DriverMySQL {
				if !strings.Contains(sql, normalizeSchema("UNIQUE INDEX endorsements_pair_key (skill_id, user_id)")) {
					t.Errorf("MySQL did not take its index inline:\n%s", sql)
				}
			} else if !strings.Contains(sql, normalizeSchema(
				"CREATE UNIQUE INDEX endorsements_pair_key ON endorsements (skill_id, user_id);")) {
				t.Errorf("%s did not get a separate unique index:\n%s", driver, sql)
			}
		})
	}
}
