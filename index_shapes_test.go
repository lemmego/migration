package migration

import (
	"fmt"
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
	// Postgres and SQLite only: the example carries a partial index, which
	// MySQL has no form of — so the README's own table says so and Validate
	// refuses it there. That refusal is covered by
	// TestAPartialIndexIsRefusedOnMySQL.
	for _, driver := range []string{DriverPostgres, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			schema := CreateFor(driver, "endorsements", func(t *Table) {
				t.UnsignedBigInt("endorser_id")
				t.UnsignedBigInt("user_id")
				t.UnsignedBigInt("skill_id")
				t.String("note", 140).Default("")
				t.TimestampTz("created_at", 0).Default(Expr("CURRENT_TIMESTAMP"))

				t.Check("endorsements_not_self", "endorser_id <> user_id")

				t.Json("payload").Nullable()

				t.Foreign("user_id", "skill_id").
					Name("endorsements_claimed_skill").
					References("user_id", "skill_id").
					On("user_skills").
					OnDelete("CASCADE")

				t.Index("user_id", "created_at").Name("endorsements_recent_idx")

				t.Index("user_id", "created_at").
					Desc("created_at").
					Name("endorsements_feed_idx")

				t.Index("skill_id", "user_id").Unique().Name("endorsements_pair_key")

				t.UniqueKey("user_id", "note").Name("endorsements_one_note")

				t.Index("user_id").
					Where("note <> ''").
					Name("endorsements_annotated_idx")
			})

			if err := schema.Validate(); err != nil {
				t.Fatalf("the README example does not validate: %v", err)
			}
			sql := normalizeSchema(schema.Build())

			// Each claim the prose makes about this example, asserted.
			for _, want := range []string{
				"CONSTRAINT endorsements_not_self CHECK (endorser_id <> user_id)",
				"CONSTRAINT endorsements_claimed_skill FOREIGN KEY (user_id, skill_id) " +
					"REFERENCES user_skills(user_id, skill_id)",
				"CONSTRAINT endorsements_one_note UNIQUE (user_id, note)",
				"(user_id, created_at DESC)",
			} {
				if !strings.Contains(sql, normalizeSchema(want)) {
					t.Errorf("expected %q in:\n%s", want, sql)
				}
			}

			if !strings.Contains(sql, normalizeSchema(
				"CREATE UNIQUE INDEX endorsements_pair_key ON endorsements (skill_id, user_id);")) {
				t.Errorf("%s did not get a separate unique index:\n%s", driver, sql)
			}
			if !strings.Contains(sql, normalizeSchema(
				"CREATE INDEX endorsements_annotated_idx ON endorsements (user_id) WHERE note <> ''")) {
				t.Errorf("%s did not get the partial index:\n%s", driver, sql)
			}
		})
	}
}

// A foreign key can be named, which is what lets an existing hand-written
// constraint be restated rather than replaced.
//
// Opt-in: without Name the key stays anonymous, because every foreign key
// this package has emitted so far was, and the server-derived name is what
// sits in databases that have already run those migrations. Emitting the
// derived name by default would rename all of them.
func TestAForeignKeyCanBeNamed(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			named := buildFor(t, driver, func(tb *Table) {
				tb.UnsignedBigInt("user_id")
				tb.Foreign("user_id").Name("posts_author_fk").
					References("id").On("users").OnDelete("CASCADE")
			})
			if !strings.Contains(named, normalizeSchema("CONSTRAINT posts_author_fk FOREIGN KEY (user_id)")) {
				t.Fatalf("the name was not emitted:\n%s", named)
			}

			anonymous := buildFor(t, driver, func(tb *Table) {
				tb.UnsignedBigInt("user_id")
				tb.Foreign("user_id").References("id").On("users").OnDelete("CASCADE")
			})
			if strings.Contains(anonymous, normalizeSchema("CONSTRAINT")) {
				t.Fatalf("an unnamed foreign key gained a name, renaming every one "+
					"that already exists:\n%s", anonymous)
			}
		})
	}
}

// A document column, which had no spelling at all — the reason an application
// with a JSONB column had to write its whole CREATE TABLE by hand.
func TestADocumentColumnCanBeDeclared(t *testing.T) {
	for driver, want := range map[string]string{
		// JSONB, not JSON: it stores a parsed form, is cheaper to read than
		// to write, and can carry a GIN index.
		DriverPostgres: "payload JSONB",
		DriverMySQL:    "payload JSON",
		// Not a fallback — SQLite stores JSON as TEXT and its JSON1
		// functions operate on TEXT.
		DriverSQLite: "payload TEXT",
	} {
		t.Run(driver, func(t *testing.T) {
			sql := buildFor(t, driver, func(tb *Table) {
				tb.BigIncrements("id").Primary()
				tb.Json("payload").Nullable()
			})
			if !strings.Contains(sql, normalizeSchema(want)) {
				t.Fatalf("expected %q in:\n%s", want, sql)
			}
		})
	}
}

// MySQL refuses a JSON column in a key outright, and no key length helps —
// so the advice has to differ from the one TEXT gets, or it reads as "store
// your document in 255 characters".
func TestIndexingADocumentColumnIsRefusedOnMySQL(t *testing.T) {
	err := CreateFor(DriverMySQL, "posts", func(tb *Table) {
		tb.Json("payload")
		tb.Index("payload").Name("posts_payload_idx")
	}).Validate()
	if err == nil {
		t.Fatal("expected indexing a JSON column on MySQL to be refused")
	}
	if !strings.Contains(err.Error(), "cannot use in an index at all") {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "String(") {
		t.Errorf("the TEXT advice leaked into the JSON case: %v", err)
	}
	// Postgres indexes one happily, so this is a MySQL restriction and not a
	// property of the type.
	if err := CreateFor(DriverPostgres, "posts", func(tb *Table) {
		tb.Json("payload")
		tb.Index("payload").Name("posts_payload_idx")
	}).Validate(); err != nil {
		t.Errorf("Postgres refused a JSON index: %v", err)
	}
}

// A partial index, on the two dialects that have one.
func TestAPartialIndexCanBeDeclared(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			sql := buildFor(t, driver, func(tb *Table) {
				tb.String("email", 255)
				tb.TimestampTz("deleted_at", 0).Nullable()
				tb.Index("email").Unique().
					Where("deleted_at IS NULL").
					Name("posts_live_email_key")
			})
			want := "CREATE UNIQUE INDEX posts_live_email_key ON posts (email) WHERE deleted_at IS NULL;"
			if !strings.Contains(sql, normalizeSchema(want)) {
				t.Fatalf("expected %q in:\n%s", want, sql)
			}
		})
	}
}

// MySQL has no partial index in any form, so one is refused rather than
// rendered without its predicate.
//
// This is the Enum lesson applied before it bites: an index silently covering
// every row instead of a tenth of them gives correct results, an unaccountable
// plan, and a table that grows an index it was promised would stay small.
func TestAPartialIndexIsRefusedOnMySQL(t *testing.T) {
	err := CreateFor(DriverMySQL, "posts", func(tb *Table) {
		tb.String("email", 255)
		tb.TimestampTz("deleted_at", 0).Nullable()
		tb.Index("email").Where("deleted_at IS NULL").Name("posts_live_idx")
	}).Validate()
	if err == nil {
		t.Fatal("expected a partial index on MySQL to be refused")
	}
	for _, want := range []string{"no partial index", "posts_live_idx"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q: %v", want, err)
		}
	}
}

// MySQL has no IF EXISTS on any drop of a constraint, a key or an index —
// not on the ALTER TABLE action and not on DROP INDEX.
//
// Refused for every kind, including an index. The first version of the check
// exempted indexes and therefore dropped the request on the floor: the
// migration runs, the tolerance is gone, and the only sign is a rollback that
// fails the second time.
func TestATolerantDropIsRefusedOnMySQL(t *testing.T) {
	for _, probe := range []struct {
		what string
		fn   func(*Table)
	}{
		{"an index", func(tb *Table) { tb.DropIndex("posts_email_idx").IfExists() }},
		{"a check", func(tb *Table) { tb.DropCheck("posts_kind_valid").IfExists() }},
		{"a foreign key", func(tb *Table) { tb.DropForeignKey("posts_author_fk").IfExists() }},
	} {
		t.Run(probe.what, func(t *testing.T) {
			err := AlterFor(DriverMySQL, "posts", probe.fn).Validate()
			if err == nil {
				t.Fatalf("expected a tolerant drop of %s to be refused on MySQL", probe.what)
			}
			if !strings.Contains(err.Error(), "no IF EXISTS") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Each drop, spelled the way the dialect spells it. These four emitted
// nothing at all before — the constraint carried only a name, and every
// branch of the drop path discriminated on a field that was nil, so the
// output was `ALTER TABLE t ;`.
func TestEachDropIsSpelledPerDialect(t *testing.T) {
	for _, probe := range []struct {
		what     string
		fn       func(*Table)
		postgres string
		mysql    string
	}{
		{"a primary key", func(tb *Table) { tb.DropPrimaryKey() },
			"ALTER TABLE posts DROP CONSTRAINT posts_pkey;",
			"ALTER TABLE posts DROP PRIMARY KEY;"},
		{"a unique key", func(tb *Table) { tb.DropUniqueKey("posts_slug_unique") },
			"ALTER TABLE posts DROP CONSTRAINT posts_slug_unique;",
			// Backed by an index on MySQL, so that is how it goes.
			"ALTER TABLE posts DROP INDEX posts_slug_unique;"},
		{"a foreign key", func(tb *Table) { tb.DropForeignKey("posts_author_fk") },
			"ALTER TABLE posts DROP CONSTRAINT posts_author_fk;",
			"ALTER TABLE posts DROP FOREIGN KEY posts_author_fk;"},
		{"an index", func(tb *Table) { tb.DropIndex("posts_email_idx") },
			// Not an ALTER TABLE action outside MySQL: an index name is
			// schema-global, so it is dropped without naming the table.
			"DROP INDEX posts_email_idx;",
			"ALTER TABLE posts DROP INDEX posts_email_idx;"},
	} {
		t.Run(probe.what, func(t *testing.T) {
			for driver, want := range map[string]string{
				DriverPostgres: probe.postgres,
				DriverMySQL:    probe.mysql,
			} {
				got := normalizeSchema(strings.Join(
					AlterFor(driver, "posts", probe.fn).Statements(), " "))
				if got != normalizeSchema(want) {
					t.Errorf("%s: expected %q, got %q", driver, want, got)
				}
			}
		})
	}
}

// The README's "Altering a table" example, compiled and rendered for each
// dialect, with the table of spellings in that section asserted rather than
// described.
func TestTheReadmeAlterExampleBuilds(t *testing.T) {
	example := func(t *Table) {
		t.Int("view_count").Default(0)
		t.DropColumn("legacy_body")
		t.RenameColumn("body", "body_source")
		t.DropIndex("posts_author_idx")
	}

	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			schema := AlterFor(driver, "posts", example)
			if err := schema.Validate(); err != nil {
				t.Fatalf("the README alter example does not validate: %v", err)
			}
			statements := schema.Statements()

			// The example is three column actions plus an index drop, and the
			// statement count is the README's last table row made concrete:
			// MySQL takes all four in one ALTER; Postgres takes the three
			// column actions in one and drops the index separately; SQLite
			// takes one action per ALTER, so three plus the drop.
			wantCount := map[string]int{
				DriverMySQL:    1,
				DriverPostgres: 2,
				DriverSQLite:   4,
			}[driver]
			if len(statements) != wantCount {
				t.Errorf("expected %d statements, got %d:\n%s",
					wantCount, len(statements), strings.Join(statements, "\n"))
			}

			// And where the index drop lands, which is the row that differs
			// three ways.
			if driver == DriverMySQL {
				if !strings.HasPrefix(normalizeSchema(statements[0]), normalizeSchema("ALTER TABLE posts")) ||
					!strings.Contains(normalizeSchema(statements[0]), normalizeSchema("DROP INDEX posts_author_idx")) {
					t.Errorf("MySQL did not drop its index inside the ALTER:\n%s", statements[0])
				}
				return
			}
			last := normalizeSchema(statements[len(statements)-1])
			if last != normalizeSchema("DROP INDEX posts_author_idx;") {
				t.Errorf("%s did not drop its index as its own statement, got %q", driver, last)
			}
		})
	}

	// And the column change, which SQLite refuses outright.
	if err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		AlterFor(DriverSQLite, "posts", func(t *Table) {
			t.String("slug", 160).Nullable().Change()
		}).Build()
		return nil
	}(); err == nil {
		t.Error("SQLite accepted a column change, which the README says it refuses")
	}
}
