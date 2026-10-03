package migration

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// postgresDB opens the database named by MIGRATION_POSTGRES_DSN, skipping when
// it is unset. SQLite and MySQL buffer a result set, so they cannot exercise
// the cursor-overlap behaviour this file exists for — only a driver that
// streams, like lib/pq, can.
func postgresDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MIGRATION_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set MIGRATION_POSTGRES_DSN to run the Postgres migration tests")
	}
	db, err := sql.Open(DriverPostgres, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("pinging %s: %v", dsn, err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS migration_down_probe`)
		_, _ = db.Exec(`DROP TABLE IF EXISTS schema_migrations`)
		db.Close()
	})
	// Start from nothing, so a previous run cannot mask a failure.
	if _, err := db.Exec(`DROP TABLE IF EXISTS migration_down_probe`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE IF EXISTS schema_migrations`); err != nil {
		t.Fatal(err)
	}
	return db
}

// Down used to run each reverted migration while the cursor that selected the
// versions was still open on the same transaction. A *sql.Tx holds one
// connection, so a statement issued mid-iteration desynchronises the wire
// protocol: lib/pq reported "unexpected Parse response 'C'" on the first
// statement the Down function ran, and `migrate down` failed every time on
// Postgres while passing on MySQL and SQLite.
func TestDownRevertsOnPostgres(t *testing.T) {
	db := postgresDB(t)

	oldMigrator := migrator
	defer func() { migrator = oldMigrator }()
	migrator = &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}

	migrator.AddMigration(&Migration{
		Version: "20260101000000",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`CREATE TABLE migration_down_probe (id BIGSERIAL PRIMARY KEY)`)
			return err
		},
		// A Down that actually executes SQL is the point: a no-op Down would
		// still hit the bug at the bookkeeping DELETE, but this reproduces
		// the failure exactly as a real migration meets it.
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`DROP TABLE migration_down_probe`)
			return err
		},
	})

	m, err := Init(db, DriverPostgres)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Up(-1); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if !tableExists(t, db, "migration_down_probe") {
		t.Fatal("Up did not create the table")
	}

	if err := m.Down(-1); err != nil {
		t.Fatalf("Down: %v", err)
	}
	if tableExists(t, db, "migration_down_probe") {
		t.Error("Down reported success but the table is still there")
	}

	var remaining int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Errorf("schema_migrations still holds %d row(s) after Down", remaining)
	}
}

// Reverting several migrations in one call iterates the cursor more than once,
// which is where the overlap was most likely to bite.
func TestDownRevertsSeveralMigrationsOnPostgres(t *testing.T) {
	db := postgresDB(t)

	oldMigrator := migrator
	defer func() { migrator = oldMigrator }()
	migrator = &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}

	for _, v := range []string{"20260101000001", "20260101000002", "20260101000003"} {
		version := v
		migrator.AddMigration(&Migration{
			Version: version,
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec(`CREATE TABLE probe_` + version + ` (id INT)`)
				return err
			},
			Down: func(tx *sql.Tx) error {
				_, err := tx.Exec(`DROP TABLE probe_` + version)
				return err
			},
		})
		t.Cleanup(func() { _, _ = db.Exec(`DROP TABLE IF EXISTS probe_` + version) })
	}

	m, err := Init(db, DriverPostgres)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Up(-1); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if err := m.Down(-1); err != nil {
		t.Fatalf("Down: %v", err)
	}
	for _, v := range []string{"20260101000001", "20260101000002", "20260101000003"} {
		if tableExists(t, db, "probe_"+v) {
			t.Errorf("probe_%s survived Down", v)
		}
	}
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1`,
		name).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n > 0
}

// The migrator is a package-level singleton and done only ever moved from
// false to true, so a process that initialised against a second database
// carried the first one's state across and skipped every migration — leaving
// an empty schema and reporting success. That is exactly what a test harness
// with a scratch database per case does, and what a program migrating one
// tenant after another does.
func TestInitTakesDoneStateFromTheDatabase(t *testing.T) {
	first := postgresDB(t)

	oldMigrator := migrator
	defer func() { migrator = oldMigrator }()
	migrator = &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}

	migrator.AddMigration(&Migration{
		Version: "20260301000000",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`CREATE TABLE tenant_probe (id INT)`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`DROP TABLE tenant_probe`)
			return err
		},
	})

	m1, err := Init(first, DriverPostgres)
	if err != nil {
		t.Fatal(err)
	}
	if err := m1.Up(-1); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, first, "tenant_probe") {
		t.Fatal("the first database was not migrated")
	}

	// A second, independent database. Same process, same migrator.
	second := secondPostgresDB(t)

	m2, err := Init(second, DriverPostgres)
	if err != nil {
		t.Fatal(err)
	}
	if err := m2.Up(-1); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, second, "tenant_probe") {
		t.Error("the second database was not migrated: Init carried done state over from the first")
	}
}

// secondPostgresDB opens a scratch database beside the one in the DSN, so the
// test has two genuinely separate schemas in one process.
func secondPostgresDB(t *testing.T) *sql.DB {
	t.Helper()
	admin, err := sql.Open(DriverPostgres, os.Getenv("MIGRATION_POSTGRES_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	name := "migration_second_probe"
	if _, err := admin.Exec(`DROP DATABASE IF EXISTS ` + name + ` WITH (FORCE)`); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(`CREATE DATABASE ` + name); err != nil {
		t.Fatal(err)
	}

	fields := map[string]string{}
	for _, field := range strings.Fields(os.Getenv("MIGRATION_POSTGRES_DSN")) {
		if k, v, ok := strings.Cut(field, "="); ok {
			fields[k] = v
		}
	}
	fields["dbname"] = name
	var pairs []string
	for k, v := range fields {
		pairs = append(pairs, k+"="+v)
	}

	db, err := sql.Open(DriverPostgres, strings.Join(pairs, " "))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		cleanup, err := sql.Open(DriverPostgres, os.Getenv("MIGRATION_POSTGRES_DSN"))
		if err == nil {
			_, _ = cleanup.Exec(`DROP DATABASE IF EXISTS ` + name + ` WITH (FORCE)`)
			cleanup.Close()
		}
	})
	return db
}
