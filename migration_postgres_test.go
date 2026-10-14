package migration

import (
	"database/sql"
	"net/url"
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
	// These tests share one database, so each run starts from nothing and
	// leaves nothing behind. Dropping every probe table rather than the one
	// the calling test knows about keeps a half-finished earlier run from
	// failing a later one with "relation already exists", which says nothing
	// about the code under test.
	reset := func() {
		for _, stmt := range []string{
			`DROP TABLE IF EXISTS migration_down_probe`,
			`DROP TABLE IF EXISTS tenant_probe`,
			`DROP TABLE IF EXISTS probe_20260101000001`,
			`DROP TABLE IF EXISTS probe_20260101000002`,
			`DROP TABLE IF EXISTS probe_20260101000003`,
			`DROP TABLE IF EXISTS schema_migrations`,
		} {
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("resetting the test database: %v", err)
			}
		}
	}
	reset()
	t.Cleanup(func() {
		reset()
		db.Close()
	})
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

// A stepped Down on Postgres, which is the only place the step's SQL actually
// runs: it is a `LIMIT $1`, and postgresDialect numbers placeholders $N while
// SQLite and MySQL take `?`. Both existing tests here call Down(-1), so the
// stepped query had never been executed against a real Postgres driver — the
// branch could have been malformed and every SQLite test would still pass.
//
// It also covers the behaviour the step change was about: three migrations in
// one batch, and Down(1) takes one of them.
func TestDownStepOnPostgresTakesOneMigrationFromOneBatch(t *testing.T) {
	db := postgresDB(t)

	oldMigrator := migrator
	defer func() { migrator = oldMigrator }()
	migrator = &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}

	versions := []string{"20260101000001", "20260101000002", "20260101000003"}
	for _, v := range versions {
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
	// One batch, three migrations.
	if err := m.Up(-1); err != nil {
		t.Fatalf("Up: %v", err)
	}

	if err := m.Down(1); err != nil {
		t.Fatalf("Down(1): %v", err)
	}

	// The newest is gone and the other two are untouched.
	if tableExists(t, db, "probe_20260101000003") {
		t.Error("Down(1) did not revert the most recent migration")
	}
	for _, v := range []string{"20260101000001", "20260101000002"} {
		if !tableExists(t, db, "probe_"+v) {
			t.Errorf("Down(1) also reverted probe_%s; a step is a migration, not a batch", v)
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

	db, err := sql.Open(DriverPostgres, dsnNaming(os.Getenv("MIGRATION_POSTGRES_DSN"), name))
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

// dsnNaming returns dsn pointed at a different database.
//
// lib/pq accepts two DSN shapes and this only understood one. The keyword
// rewrite it used to do — split on whitespace, cut on "=", put the pairs back
// together — turns a URL into nonsense rather than refusing it: "sslmode"
// becomes part of a key and the setting is lost, so the second connection
// went out with lib/pq's own default of require and the test failed with
// "pq: SSL is not enabled on the server". That names the symptom and hides
// the cause, which is a confusing thing to be told when the DSN you supplied
// was fine.
func dsnNaming(dsn, database string) string {
	if parsed, err := url.Parse(dsn); err == nil &&
		(parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		parsed.Path = "/" + database
		return parsed.String()
	}

	fields := map[string]string{}
	for _, field := range strings.Fields(dsn) {
		if key, value, ok := strings.Cut(field, "="); ok {
			fields[key] = value
		}
	}
	fields["dbname"] = database

	pairs := make([]string, 0, len(fields))
	for key, value := range fields {
		pairs = append(pairs, key+"="+value)
	}
	return strings.Join(pairs, " ")
}
