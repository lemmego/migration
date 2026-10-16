package migration

import (
	"os"
	"strings"
	"testing"
)

// The SQLite test proves the statement parses; this proves it parses and
// enforces on the dialect the constraint was added for. Postgres is also the
// only one of the three where the index-name collision is real, since index
// names are schema-global there.
func TestChecksAndIndexNamesOnRealPostgres(t *testing.T) {
	db := postgresDB(t) // skips without MIGRATION_POSTGRES_DSN
	os.Setenv("DB_DRIVER", DriverPostgres)

	for _, table := range []string{"check_probe_a", "check_probe_b"} {
		schema := Create(table, func(tb *Table) {
			tb.Increments("id").Primary()
			tb.BigInt("user_id")
			tb.Int("position")
			tb.Enum("state", "draft", "published")
			tb.Check(table+"_position_sane", "position >= 0")
			// The collision case: both tables want this pair.
			tb.Index("user_id", "position")
		}).Build()
		t.Cleanup(func() { _, _ = db.Exec(`DROP TABLE IF EXISTS ` + table) })

		for _, statement := range splitStatements(schema) {
			if _, err := db.Exec(statement); err != nil {
				t.Fatalf("Postgres refused:\n%s\n\n%v", statement, err)
			}
		}
	}

	// The enum CHECK is enforced, not merely present in the text.
	if _, err := db.Exec(`INSERT INTO check_probe_a (user_id, position, state) VALUES (1, 0, 'draft')`); err != nil {
		t.Fatalf("a permitted row was refused: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO check_probe_a (user_id, position, state) VALUES (1, 0, 'nonsense')`); err == nil {
		t.Error("Enum accepted a value outside its list, so it constrains nothing")
	}
	if _, err := db.Exec(`INSERT INTO check_probe_a (user_id, position, state) VALUES (1, -1, 'draft')`); err == nil {
		t.Error("the explicit Check accepted a row it forbids")
	}
}

// splitStatements splits the blob Build returns. Statements() already does
// this properly; this keeps the test honest about executing each one.
func splitStatements(schema string) []string {
	var out []string
	for _, part := range strings.Split(schema, ";") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
