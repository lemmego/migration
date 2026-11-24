package migration

import (
	"errors"
	"strings"
	"testing"
)

// An unrecognised dialect must be refused, and the regression is specific:
// buildCreate and buildDrop had no default branch and returned "", while
// Statements went on appending droppedIndexes and trailingIndexes. So a
// migration emitted CREATE INDEX against a table it had never created, and
// reported success.
//
// Statements is asserted rather than Build because Statements is where the
// orphan was visible: Build joins the same slice, so a non-empty Build could
// have been the index alone.
func TestAnUnknownDialectIsRefusedRatherThanEmittingAnOrphanedIndex(t *testing.T) {
	schema := CreateFor("nonsense", "widgets", func(table *Table) {
		table.String("name", 255)
		table.Index("name")
	})

	err := schema.Validate()
	if err == nil {
		t.Fatal("Validate accepted an unknown dialect")
	}
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Errorf("error does not wrap ErrUnsupportedDialect: %v", err)
	}
	if !strings.Contains(err.Error(), "nonsense") {
		t.Errorf("the error does not name the dialect it was given: %v", err)
	}

	// And the orphan is unreachable, because Statements validates first.
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("Statements returned instead of panicking on an unknown dialect")
		}
		// The old behaviour: a CREATE INDEX with no CREATE TABLE.
		if statements, ok := recovered.([]string); ok {
			t.Fatalf("Statements produced %v", statements)
		}
	}()
	statements := schema.Statements()
	for _, statement := range statements {
		if strings.Contains(strings.ToUpper(statement), "CREATE INDEX") {
			t.Fatalf("the orphaned index is back: %q", statement)
		}
	}
}

// A Drop carries no table, and Validate returned nil early for exactly that
// case — so the dialect check has to come before it or every drop skips the
// check entirely.
func TestAnUnknownDialectIsRefusedForADropToo(t *testing.T) {
	err := DropFor("nonsense", "widgets").Validate()
	if err == nil {
		t.Fatal("a drop with an unknown dialect was accepted; the check is after the table == nil return")
	}
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Errorf("error does not wrap ErrUnsupportedDialect: %v", err)
	}
}

// The three real dialects and the spellings people actually write.
func TestNormaliseDialectAcceptsTheSpellingsPeopleWrite(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"sqlite", DriverSQLite},
		{"sqlite3", DriverSQLite},
		{"libsql", DriverSQLite}, // a fork of SQLite; same DDL
		{"LibSQL", DriverSQLite},
		{"  sqlite3  ", DriverSQLite},
		{"mysql", DriverMySQL},
		{"mariadb", DriverMySQL},
		{"postgres", DriverPostgres},
		{"postgresql", DriverPostgres},
		{"pgsql", DriverPostgres},
		{"pgx", DriverPostgres},
	} {
		got, ok := NormaliseDialect(c.in)
		if !ok {
			t.Errorf("NormaliseDialect(%q) was not recognised", c.in)
			continue
		}
		if got != c.want {
			t.Errorf("NormaliseDialect(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	for _, bad := range []string{"", "nonsense", "mssql", "sqlserver", "mongo"} {
		if _, ok := NormaliseDialect(bad); ok {
			t.Errorf("NormaliseDialect(%q) was accepted", bad)
		}
	}
}

// DB_DRIVER unset is the one case that legitimately defaults. A value that is
// set and unrecognised must NOT default, because defaulting is the other half
// of the original bug: a typo generated SQLite DDL for a Postgres database and
// the value never appeared anywhere.
func TestDetermineDialectDefaultsOnlyWhenTheVariableIsUnset(t *testing.T) {
	t.Setenv("DB_DRIVER", "")
	if got := determineDialect(); got != DriverSQLite {
		t.Errorf("unset DB_DRIVER gave %q, want %q", got, DriverSQLite)
	}

	t.Setenv("DB_DRIVER", "pgx")
	if got := determineDialect(); got != DriverPostgres {
		t.Errorf("DB_DRIVER=pgx gave %q, want %q", got, DriverPostgres)
	}

	t.Setenv("DB_DRIVER", "nonsense")
	if got := determineDialect(); got == DriverSQLite {
		t.Error("a typo'd DB_DRIVER silently became sqlite, which is the bug")
	} else if got != "nonsense" {
		t.Errorf("determineDialect returned %q; it should hand the raw value through so Validate can name it", got)
	}
}

// DataSource must keep refusing libsql, even though NormaliseDialect accepts
// it. A libSQL server's address is a URL, and building a DSN for it here would
// produce a sqlite FILE path — silently opening a local file instead of
// connecting to the server, which is worse than the refusal.
func TestTheDSNBuilderStillRefusesLibSQL(t *testing.T) {
	_, err := (&DataSource{Driver: "libsql", Name: "app"}).String()
	if err == nil {
		t.Fatal("DataSource built a DSN for libsql; it would be a local file path")
	}
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Errorf("error does not wrap ErrUnsupportedDialect: %v", err)
	}
}
