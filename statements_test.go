package migration

import (
	"strings"
	"testing"
)

// Build joins its statements with a newline, which a driver accepting several
// at once will run. MySQL's driver does not, unless the DSN opts in, so a
// caller that executes DDL itself needs them apart — and splitting Build's
// output on semicolons afterwards is guesswork.
func TestStatementsSeparatesIndexesFromTheTable(t *testing.T) {
	for _, dialect := range []string{DriverSQLite, DriverPostgres} {
		schema := CreateFor(dialect, "widgets", func(tb *Table) {
			tb.String("id", 64).Primary()
			tb.String("owner", 64)
			tb.Index("owner")
		})

		statements := schema.Statements()
		if len(statements) != 2 {
			t.Fatalf("%s: got %d statements, want a table and an index:\n%v",
				dialect, len(statements), statements)
		}
		if !strings.HasPrefix(statements[0], "CREATE TABLE") {
			t.Errorf("%s: first statement is not the table:\n%s", dialect, statements[0])
		}
		if !strings.HasPrefix(statements[1], "CREATE INDEX") {
			t.Errorf("%s: second statement is not the index:\n%s", dialect, statements[1])
		}
		for i, statement := range statements {
			if strings.Contains(strings.TrimSuffix(strings.TrimSpace(statement), ";"), ";") {
				t.Errorf("%s: statement %d holds more than one statement:\n%s", dialect, i, statement)
			}
		}
	}
}

// MySQL declares an index inside CREATE TABLE, so there is nothing trailing
// and the whole schema really is one statement.
func TestStatementsIsASingleStatementOnMySQL(t *testing.T) {
	schema := CreateFor(DriverMySQL, "widgets", func(tb *Table) {
		tb.String("id", 64).Primary()
		tb.String("owner", 64)
		tb.Index("owner")
	})
	if statements := schema.Statements(); len(statements) != 1 {
		t.Fatalf("got %d statements, want 1:\n%v", len(statements), statements)
	}
}

// Build and Statements must describe the same schema, or a caller picking one
// would get different DDL from the other.
func TestBuildAndStatementsAgree(t *testing.T) {
	for _, dialect := range []string{DriverSQLite, DriverMySQL, DriverPostgres} {
		build := func() *Schema {
			return CreateFor(dialect, "widgets", func(tb *Table) {
				tb.String("id", 64).Primary()
				tb.String("owner", 64)
				tb.Index("owner")
			})
		}
		joined := strings.Join(build().Statements(), "\n")
		if got := build().Build(); got != joined {
			t.Errorf("%s: Build and Statements disagree\nBuild:\n%s\nStatements:\n%s", dialect, got, joined)
		}
	}
}

// The dialect has to come from the argument, not from DB_DRIVER, or a package
// emitting DDL for a connection it holds would get whatever the environment
// happened to say.
func TestCreateForIgnoresTheEnvironment(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverMySQL)

	postgres := CreateFor(DriverPostgres, "widgets", func(tb *Table) {
		tb.DateTime("created_at", 6)
	}).Build()

	if strings.Contains(postgres, "DATETIME") {
		t.Errorf("CreateFor(postgres) emitted MySQL types while DB_DRIVER said mysql:\n%s", postgres)
	}
	if !strings.Contains(strings.ToUpper(postgres), "TIMESTAMP") {
		t.Errorf("CreateFor(postgres) did not emit a postgres timestamp:\n%s", postgres)
	}
}

// A package shipping a migration renders its own DDL and has to know which
// database it is targeting; the schema builder's DB_DRIVER default is not
// visible to it.
func TestMigratorReportsItsDialect(t *testing.T) {
	m := GetMigrator()
	if got := m.Dialect(); got != "" && got != DriverSQLite && got != DriverMySQL && got != DriverPostgres {
		t.Fatalf("Dialect() = %q, want empty or a known driver", got)
	}

	previous := m.dialect
	t.Cleanup(func() { m.dialect = previous })

	m.dialect = DriverPostgres
	if got := m.Dialect(); got != DriverPostgres {
		t.Errorf("Dialect() = %q, want %q", got, DriverPostgres)
	}
}
