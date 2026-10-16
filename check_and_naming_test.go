package migration

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/glebarez/go-sqlite"
)

func sql2open(t *testing.T) (*sql.DB, error) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err == nil {
		t.Cleanup(func() { db.Close() })
	}
	return db, err
}

func build(t *testing.T, driver string, fn func(*Table)) string {
	t.Helper()
	os.Setenv("DB_DRIVER", driver)
	return normalizeSchema(Create("posts", fn).Build())
}

// A CHECK constraint is standard SQL — Postgres, MySQL 8.0.16+ and SQLite all
// enforce one — and the builder had no way to express it. An application whose
// columns carry states, ranges or formats then writes every one of them as raw
// tx.Exec, which is how a schema ends up more hand-written than generated.
func TestACheckConstraintCanBeDeclared(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		t.Run(driver, func(t *testing.T) {
			sql := build(t, driver, func(tb *Table) {
				tb.Increments("id").Primary()
				tb.String("state", 32)
				tb.Check("posts_state_valid", "state IN ('draft', 'published')")
			})
			want := normalizeSchema("CONSTRAINT posts_state_valid CHECK (state IN ('draft', 'published'))")
			if !strings.Contains(sql, want) {
				t.Errorf("CREATE TABLE did not carry the check\n want: %s\n got:\n%s", want, sql)
			}
		})
	}
}

// Postgres index names are schema-global, so two tables each wanting an index
// on (user_id, position) collided — and because the whole migration batch runs
// in one transaction, the second one took the batch down with it. The name was
// built from the columns alone for the multi-column case, where the
// single-column case had the table prefix all along.
func TestAMultiColumnIndexNameCarriesItsTable(t *testing.T) {
	sql := build(t, DriverPostgres, func(tb *Table) {
		tb.Increments("id").Primary()
		tb.BigInt("user_id")
		tb.Int("position")
		tb.Index("user_id", "position")
	})

	if strings.Contains(sql, normalizeSchema("CREATE INDEX user_id_position_index")) {
		t.Fatal("the index name has no table in it, so a second table wanting " +
			"(user_id, position) collides and fails the batch")
	}
	if !strings.Contains(sql, normalizeSchema("CREATE INDEX posts_user_id_position_index")) {
		t.Errorf("expected a table-prefixed index name, got:\n%s", sql)
	}
}

// A unique key creates an index in Postgres, so its name is schema-global too
// — and here the bug was inverted: the single-column branch overwrote the
// prefix it had just built, so `email_unique` was claimed by whichever table
// migrated first.
func TestASingleColumnUniqueKeyNameCarriesItsTable(t *testing.T) {
	sql := build(t, DriverMySQL, func(tb *Table) {
		tb.Increments("id").Primary()
		tb.String("email", 255)
		tb.UniqueKey("email")
	})

	if strings.Contains(sql, normalizeSchema("UNIQUE email_unique (")) {
		t.Fatal("the unique constraint name has no table in it")
	}
	if !strings.Contains(sql, "posts_email_unique") {
		t.Errorf("expected a table-prefixed unique name, got:\n%s", sql)
	}
}

// Enum emitted a bare column on Postgres and enforced nothing.
//
// NewDataType calls AddSuffixes before Enum chains WithEnumValues on, so
// enumValues was still nil when the CHECK would have been appended. And had
// the order been right, the values were formatted with %s on a []string, which
// renders `IN ([draft published])` — unquoted, space-separated, invalid.
//
// Silence is the bad half: a column declared as an enum that accepts any
// string at all is worse than no column, because the schema says it is
// constrained.
func TestEnumIsConstrainedOnPostgres(t *testing.T) {
	sql := build(t, DriverPostgres, func(tb *Table) {
		tb.Increments("id").Primary()
		tb.Enum("state", "draft", "published")
	})

	want := normalizeSchema("CHECK (state IN ('draft', 'published'))")
	if !strings.Contains(sql, want) {
		t.Errorf("Enum enforces nothing on Postgres\n want: %s\n got:\n%s", want, sql)
	}
	if strings.Contains(sql, normalizeSchema("[draft published]")) {
		t.Error("the values were rendered as a Go slice")
	}
}

// MySQL has a native ENUM type, and it was handed the same Go slice.
func TestEnumRendersItsValuesOnMySQL(t *testing.T) {
	sql := build(t, DriverMySQL, func(tb *Table) {
		tb.Increments("id").Primary()
		tb.Enum("state", "draft", "published")
	})

	if !strings.Contains(sql, normalizeSchema("ENUM('draft', 'published')")) {
		t.Errorf("expected a rendered ENUM type, got:\n%s", sql)
	}
	if strings.Contains(sql, normalizeSchema("[draft published]")) {
		t.Error("the values were rendered as a Go slice")
	}
}

// A value with an apostrophe must not end the string literal.
func TestAnEnumValueWithAQuoteIsEscaped(t *testing.T) {
	sql := build(t, DriverPostgres, func(tb *Table) {
		tb.Increments("id").Primary()
		tb.Enum("mood", "it's fine")
	})

	if !strings.Contains(sql, normalizeSchema("'it''s fine'")) {
		t.Errorf("the quote was not doubled, so the literal ends early:\n%s", sql)
	}
}

// String comparison says the constraint is in the output; only the database
// says the statement parses. SQLite trimmed the separator after the last
// column whenever there was no unique key and no foreign key to follow — and
// a CHECK is neither, so the column ran straight into CONSTRAINT with nothing
// between them. That is the trap a new constraint kind walks into: the code
// deciding whether a separator is needed enumerated the kinds that existed
// when it was written.
func TestTheGeneratedSQLiteSchemaActuallyParses(t *testing.T) {
	os.Setenv("DB_DRIVER", DriverSQLite)
	sql := Create("posts", func(tb *Table) {
		tb.Increments("id").Primary()
		tb.String("state", 32)
		tb.Check("posts_state_valid", "state IN ('draft', 'published')")
	}).Build()

	db, err := sql2open(t)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(sql); err != nil {
		t.Fatalf("SQLite refused the generated schema: %v\n\n%s", err, sql)
	}

	if _, err := db.Exec(`INSERT INTO posts (state) VALUES ('draft')`); err != nil {
		t.Fatalf("a permitted value was refused: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO posts (state) VALUES ('nonsense')`); err == nil {
		t.Error("the check accepted a value outside its list, so it is not enforcing")
	}
}

// Adding a check to a table that already exists is at least as common as
// declaring one with it — it is what every "now constrain this column"
// migration does. Inside CREATE TABLE a constraint is part of the definition;
// in an ALTER it is an action and needs the verb, and without it the builder
// emitted `ALTER TABLE t CONSTRAINT ...`, which no dialect accepts.
func TestACheckCanBeAddedToAnExistingTable(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL} {
		t.Run(driver, func(t *testing.T) {
			os.Setenv("DB_DRIVER", driver)
			sql := normalizeSchema(Alter("posts", func(tb *Table) {
				tb.Check("posts_state_valid", "state IN ('draft', 'published')")
			}).Build())

			want := normalizeSchema("ALTER TABLE posts ADD CONSTRAINT posts_state_valid CHECK (state IN ('draft', 'published'))")
			if !strings.Contains(sql, want) {
				t.Errorf("\n want: %s\n got:  %s", want, sql)
			}
		})
	}
}

func TestACheckCanBeDroppedFromAnExistingTable(t *testing.T) {
	os.Setenv("DB_DRIVER", DriverPostgres)
	sql := normalizeSchema(Alter("posts", func(tb *Table) {
		tb.DropCheck("posts_state_valid")
	}).Build())

	if !strings.Contains(sql, normalizeSchema("DROP CONSTRAINT posts_state_valid")) {
		t.Errorf("got: %s", sql)
	}
}

// SQLite has no ALTER TABLE ADD CONSTRAINT: changing one means rebuilding the
// table. Refusing loudly beats emitting a statement the database rejects, and
// beats silently dropping the constraint — which is the failure this whole
// change exists to remove.
func TestAlteringAConstraintOnSQLiteRefuses(t *testing.T) {
	os.Setenv("DB_DRIVER", DriverSQLite)
	defer func() {
		if r := recover(); r == nil {
			t.Error("SQLite accepted an ALTER that adds a CHECK; it cannot execute one")
		} else if !strings.Contains(r.(string), "SQLite cannot alter") {
			t.Errorf("unhelpful panic: %v", r)
		}
	}()

	Alter("posts", func(tb *Table) {
		tb.Check("posts_state_valid", "state IN ('draft')")
	}).Build()
}
