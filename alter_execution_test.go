package migration

import (
	"database/sql"
	"os"
	"strings"
	"testing"
)

// Every statement the alter path emits, executed.
//
// This file exists because the bugs it covers were all of one kind: output
// that reads correctly and no server accepts. Four drop helpers emitted
// `ALTER TABLE t ;`, and a column change emitted `ALTER COLUMN x VARCHAR(n)`
// — both syntax errors, both pinned by tests that compared strings. A test
// that only compares strings cannot tell valid DDL from invalid DDL, so these
// run it.

func execAll(t *testing.T, db *sql.DB, statements []string) {
	t.Helper()
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("%s\n  -> %v", statement, err)
		}
	}
}

func TestTheAlterPathExecutesOnPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATION_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("MIGRATION_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("opening postgres: %v", err)
	}
	// t.Cleanup rather than defer, and registered first so it runs last: a
	// deferred Close runs *before* every t.Cleanup, which left the drops
	// below executing against a closed connection and the probe tables in
	// the database for the next run to collide with.
	t.Cleanup(func() { db.Close() })

	table := "alter_exec_probe"
	parent := "alter_exec_parent"
	for _, statement := range []string{
		`DROP TABLE IF EXISTS ` + table,
		`DROP TABLE IF EXISTS ` + parent,
		`DROP TABLE IF EXISTS alter_exec_idx`,
		`CREATE TABLE ` + parent + ` (id bigserial PRIMARY KEY)`,
		`CREATE TABLE ` + table + ` (
			id bigserial CONSTRAINT ` + table + `_pkey PRIMARY KEY,
			parent_id bigint,
			email varchar(255) NOT NULL,
			name varchar(50),
			CONSTRAINT ` + table + `_email_unique UNIQUE (email),
			CONSTRAINT ` + table + `_name_ok CHECK (length(name) >= 0),
			CONSTRAINT ` + table + `_parent_fk FOREIGN KEY (parent_id)
				REFERENCES ` + parent + ` (id))`,
		`CREATE INDEX ` + table + `_email_idx ON ` + table + ` (email)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("setting up: %s -> %v", statement, err)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DROP TABLE IF EXISTS ` + table)
		db.Exec(`DROP TABLE IF EXISTS ` + parent)
	})

	// A partial index, a descending key and a unique index, on a real server.
	execAll(t, db, CreateFor(DriverPostgres, "alter_exec_idx", func(tb *Table) {
		tb.BigIncrements("id").Primary()
		tb.String("email", 255)
		tb.TimestampTz("created_at", 0).Nullable()
		tb.TimestampTz("deleted_at", 0).Nullable()
		tb.Json("payload").Nullable()
		tb.Index("email", "created_at").Desc("created_at").Name("aei_recent_idx")
		tb.Index("email").Unique().Where("deleted_at IS NULL").Name("aei_live_key")
	}).Statements())
	t.Cleanup(func() { db.Exec(`DROP TABLE IF EXISTS alter_exec_idx`) })

	// Each drop, one statement at a time, so a failure names the helper.
	for _, probe := range []struct {
		what string
		fn   func(*Table)
	}{
		{"DropIndex", func(tb *Table) { tb.DropIndex(table + "_email_idx") }},
		{"DropIndex IfExists (absent)", func(tb *Table) { tb.DropIndex("aei_not_there").IfExists() }},
		{"DropForeignKey", func(tb *Table) { tb.DropForeignKey(table + "_parent_fk") }},
		{"DropUniqueKey", func(tb *Table) { tb.DropUniqueKey(table + "_email_unique") }},
		{"DropCheck IfExists", func(tb *Table) { tb.DropCheck(table + "_name_ok").IfExists() }},
		{"DropCheck IfExists (absent)", func(tb *Table) { tb.DropCheck("never_existed").IfExists() }},
		{"DropPrimaryKey", func(tb *Table) { tb.DropPrimaryKey() }},
		{"Change a column", func(tb *Table) { tb.String("name", 120).Nullable().Change() }},
		{"two columns at once", func(tb *Table) {
			tb.Int("a").Default(0)
			tb.Int("b").Default(0)
		}},
	} {
		t.Run(probe.what, func(t *testing.T) {
			execAll(t, db, AlterFor(DriverPostgres, table, probe.fn).Statements())
		})
	}
}

func TestTheAlterPathExecutesOnSQLite(t *testing.T) {
	db, err := sql2open(t)
	if err != nil {
		t.Skipf("no sqlite: %v", err)
	}

	// What SQLite can express: a table with a partial unique index and a
	// descending key, then adding and dropping columns and the index.
	execAll(t, db, CreateFor(DriverSQLite, "probe", func(tb *Table) {
		tb.BigIncrements("id").Primary()
		tb.String("email", 255)
		tb.Text("created_at").Nullable()
		tb.Text("deleted_at").Nullable()
		tb.Json("payload").Nullable()
		tb.Index("email", "created_at").Desc("created_at").Name("probe_recent_idx")
		tb.Index("email").Unique().Where("deleted_at IS NULL").Name("probe_live_key")
	}).Statements())

	for _, probe := range []struct {
		what string
		fn   func(*Table)
	}{
		{"add two columns", func(tb *Table) {
			tb.Int("a").Default(0)
			tb.Int("b").Default(0)
		}},
		{"DropColumn", func(tb *Table) { tb.DropColumn("a") }},
		{"DropIndex", func(tb *Table) { tb.DropIndex("probe_recent_idx") }},
		{"DropIndex IfExists (absent)", func(tb *Table) { tb.DropIndex("nope").IfExists() }},
	} {
		t.Run(probe.what, func(t *testing.T) {
			execAll(t, db, AlterFor(DriverSQLite, "probe", probe.fn).Statements())
		})
	}
}

// An Alter with nothing to do emits nothing, rather than `ALTER TABLE t ;`.
func TestAnEmptyAlterEmitsNothing(t *testing.T) {
	for _, driver := range []string{DriverPostgres, DriverMySQL, DriverSQLite} {
		statements := AlterFor(driver, "users", func(tb *Table) {}).Statements()
		for _, statement := range statements {
			if strings.TrimSpace(statement) != "" {
				t.Errorf("%s: an empty alter emitted %q", driver, statement)
			}
		}
	}
}
