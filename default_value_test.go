package migration

import (
	"strings"
	"testing"
	"time"
)

// A default is the value the caller wants, not a fragment of SQL they have to
// quote themselves. Default("active") must produce DEFAULT 'active'.
func TestDefaultQuotesStringsWithoutCallerHelp(t *testing.T) {
	for _, dialect := range []string{DriverSQLite, DriverMySQL, DriverPostgres} {
		t.Run(dialect, func(t *testing.T) {
			t.Setenv("DB_DRIVER", dialect)
			sql := Create("users", func(tb *Table) {
				tb.String("status", 255).Default("active")
			}).Build()

			if !strings.Contains(sql, "DEFAULT 'active'") {
				t.Errorf("expected DEFAULT 'active', got:\n%s", sql)
			}
		})
	}
}

// A quote inside the value is escaped by doubling, which every dialect accepts
// and which NO_BACKSLASH_ESCAPES does not affect.
func TestDefaultEscapesEmbeddedQuotes(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverPostgres)
	sql := Create("users", func(tb *Table) {
		tb.String("note", 255).Default("it's fine")
	}).Build()

	if !strings.Contains(sql, "DEFAULT 'it''s fine'") {
		t.Errorf("expected a doubled quote, got:\n%s", sql)
	}
}

func TestDefaultLiteralsByType(t *testing.T) {
	stamp := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)

	tests := []struct {
		name    string
		dialect string
		value   any
		want    string
	}{
		{"int", DriverSQLite, 18, "18"},
		{"negative int", DriverSQLite, -3, "-3"},
		{"uint", DriverSQLite, uint8(7), "7"},
		{"float keeps decimal form", DriverSQLite, 1.5, "1.5"},
		{"float avoids exponent", DriverSQLite, 0.00001, "0.00001"},
		{"nil is NULL", DriverSQLite, nil, "NULL"},
		{"expr passes through", DriverSQLite, Expr("CURRENT_TIMESTAMP"), "CURRENT_TIMESTAMP"},
		{"time", DriverMySQL, stamp, "'2026-03-04 05:06:07'"},

		// PostgreSQL has a real boolean type and rejects 1/0 for it; SQLite
		// and MySQL store booleans as integers.
		{"bool on postgres", DriverPostgres, true, "TRUE"},
		{"bool on mysql", DriverMySQL, true, "1"},
		{"bool on sqlite", DriverSQLite, false, "0"},

		{"bytes on mysql", DriverMySQL, []byte{0xde, 0xad}, "X'dead'"},
		{"bytes on postgres", DriverPostgres, []byte{0xde, 0xad}, `'\xdead'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultLiteral(tt.dialect, tt.value)
			if err != nil {
				t.Fatalf("defaultLiteral(%v) error = %v", tt.value, err)
			}
			if got != tt.want {
				t.Errorf("defaultLiteral(%v) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

// Sub-second precision is carried through, but a whole second does not gain a
// spurious fractional part.
func TestDefaultTimePrecision(t *testing.T) {
	withMicros := time.Date(2026, 3, 4, 5, 6, 7, 123456000, time.UTC)
	got, err := defaultLiteral(DriverPostgres, withMicros)
	if err != nil {
		t.Fatal(err)
	}
	if got != "'2026-03-04 05:06:07.123456'" {
		t.Errorf("got %s", got)
	}
}

// Default(nil) means DEFAULT NULL. Before the column carried a hasDefault
// flag, a nil value was indistinguishable from no default at all.
func TestDefaultNilEmitsNull(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverSQLite)
	sql := Create("users", func(tb *Table) {
		tb.String("nickname", 255).Nullable().Default(nil)
	}).Build()

	if !strings.Contains(sql, "DEFAULT NULL") {
		t.Errorf("expected DEFAULT NULL, got:\n%s", sql)
	}
}

// A column with no Default call must not gain a DEFAULT clause.
func TestNoDefaultEmitsNoClause(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverSQLite)
	sql := Create("users", func(tb *Table) {
		tb.String("email", 255)
	}).Build()

	if strings.Contains(sql, "DEFAULT") {
		t.Errorf("unexpected DEFAULT clause:\n%s", sql)
	}
}

// A type with no SQL literal form is reported rather than silently dropped or
// spliced in as a Go fmt rendering.
func TestDefaultRejectsUnrenderableTypes(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverSQLite)
	type point struct{ X, Y int }

	schema := Create("users", func(tb *Table) {
		tb.String("origin", 255).Default(point{1, 2})
	})

	err := schema.Validate()
	if err == nil {
		t.Fatal("expected Validate to reject a struct default")
	}
	if !strings.Contains(err.Error(), "users.origin") || !strings.Contains(err.Error(), "migration.Expr") {
		t.Errorf("error should name the column and the escape hatch, got: %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("expected Build to panic on an invalid default")
		}
	}()
	schema.Build()
}

func TestDefaultRejectsNonFiniteFloats(t *testing.T) {
	for _, value := range []float64{
		func() float64 { var z float64; return 1 / z }(),
		func() float64 { var z float64; return z / z }(),
	} {
		if _, err := defaultLiteral(DriverSQLite, value); err == nil {
			t.Errorf("expected %v to be rejected as a default", value)
		}
	}
}

// Expr is the documented escape hatch for expressions, which must not be
// quoted into a string literal.
func TestExprDefaultIsNotQuoted(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverPostgres)
	sql := Create("posts", func(tb *Table) {
		tb.DateTime("created_at", 6).Default(Expr("CURRENT_TIMESTAMP"))
	}).Build()

	if !strings.Contains(sql, "DEFAULT CURRENT_TIMESTAMP") {
		t.Errorf("expected an unquoted expression, got:\n%s", sql)
	}
	if strings.Contains(sql, "'CURRENT_TIMESTAMP'") {
		t.Errorf("expression was quoted as a literal:\n%s", sql)
	}
}

func TestValidateStillReportsUnindexableMySQLColumns(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverMySQL)
	err := Create("posts", func(tb *Table) {
		tb.Text("body").Unique()
	}).Validate()

	if err == nil {
		t.Fatal("expected the MySQL text-index check to still fire")
	}
	if !strings.Contains(err.Error(), "posts.body") {
		t.Errorf("unexpected error: %v", err)
	}
}

// CurrentTimestamp must work on every dialect without the caller knowing
// MySQL's rule that the default's precision matches the column's.
func TestCurrentTimestampDefaultPerDialect(t *testing.T) {
	tests := []struct {
		dialect string
		want    string
	}{
		{DriverSQLite, "DEFAULT CURRENT_TIMESTAMP"},
		{DriverPostgres, "DEFAULT CURRENT_TIMESTAMP"},
		{DriverMySQL, "DEFAULT CURRENT_TIMESTAMP(6)"},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			t.Setenv("DB_DRIVER", tt.dialect)
			sql := Create("posts", func(tb *Table) {
				tb.DateTime("created_at", 6).Default(CurrentTimestamp)
			}).Build()

			if !strings.Contains(sql, tt.want) {
				t.Errorf("expected %q, got:\n%s", tt.want, sql)
			}
		})
	}
}

// A column with no precision gets the bare form even on MySQL, since
// CURRENT_TIMESTAMP(0) is not the same declaration as CURRENT_TIMESTAMP.
func TestCurrentTimestampWithoutPrecision(t *testing.T) {
	t.Setenv("DB_DRIVER", DriverMySQL)
	sql := Create("posts", func(tb *Table) {
		tb.DateTime("created_at", 0).Default(CurrentTimestamp)
	}).Build()

	if !strings.Contains(sql, "DEFAULT CURRENT_TIMESTAMP") || strings.Contains(sql, "CURRENT_TIMESTAMP(") {
		t.Errorf("expected the bare form, got:\n%s", sql)
	}
}
