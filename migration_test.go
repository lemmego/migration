package migration

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/glebarez/go-sqlite"
)

func TestDownStepRollsBackLatestBatches(t *testing.T) {
	db, err := sql.Open(DriverSQLite, filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	oldMigrator := migrator
	defer func() { migrator = oldMigrator }()
	migrator = &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}

	mg1 := &Migration{Version: "001", Up: func(*sql.Tx) error { return nil }, Down: func(*sql.Tx) error { return nil }}
	mg2 := &Migration{Version: "002", Up: func(*sql.Tx) error { return nil }, Down: func(*sql.Tx) error { return nil }}
	migrator.AddMigration(mg1)
	migrator.AddMigration(mg2)
	m, err := Init(db, DriverSQLite)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(1); err != nil {
		t.Fatal(err)
	}
	if err := m.Up(1); err != nil {
		t.Fatal(err)
	}
	if err := m.Down(1); err != nil {
		t.Fatal(err)
	}

	var remaining int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("remaining migrations = %d, want 1", remaining)
	}
	if mg2.done {
		t.Fatal("latest migration is still marked done")
	}
}

func newSQLiteMigrator(t *testing.T) (*sql.DB, *Migrator) {
	t.Helper()

	db, err := sql.Open(DriverSQLite, filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	m := &Migrator{Versions: []string{}, Migrations: map[string]*Migration{}}
	oldMigrator := migrator
	migrator = m
	t.Cleanup(func() { migrator = oldMigrator })
	initialized, err := Init(db, DriverSQLite)
	if err != nil {
		t.Fatal(err)
	}
	return db, initialized
}

func TestUpRollsBackMigrationErrorAndDoneState(t *testing.T) {
	db, m := newSQLiteMigrator(t)
	wantErr := errors.New("up failed")
	mg := &Migration{
		Version: "001",
		Up: func(tx *sql.Tx) error {
			if _, err := tx.Exec("CREATE TABLE rolled_back (id INTEGER)"); err != nil {
				return err
			}
			return wantErr
		},
		Down: func(*sql.Tx) error { return nil },
	}
	m.AddMigration(mg)

	if err := m.Up(0); !errors.Is(err, wantErr) {
		t.Fatalf("Up error = %v, want %v", err, wantErr)
	}
	if mg.done {
		t.Fatal("failed migration is marked done")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("schema_migrations rows = %d, want 0", count)
	}
	if _, err := db.Exec("SELECT * FROM rolled_back"); err == nil {
		t.Fatal("migration DDL was not rolled back")
	}
}

func TestUpCommitErrorRollsBackAndPreservesDoneState(t *testing.T) {
	db, m := newSQLiteMigrator(t)
	mg := &Migration{
		Version: "001",
		Up: func(tx *sql.Tx) error {
			// Force the runner's commit path to observe a closed transaction.
			return tx.Rollback()
		},
		Down: func(*sql.Tx) error { return nil },
	}
	m.AddMigration(mg)

	if err := m.Up(0); err == nil {
		t.Fatal("Up succeeded, want transaction error")
	}
	if mg.done {
		t.Fatal("commit-failed migration is marked done")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("schema_migrations rows = %d, want 0", count)
	}
}

func TestDownRollsBackMigrationErrorAndDoneState(t *testing.T) {
	db, m := newSQLiteMigrator(t)
	mg := &Migration{
		Version: "001",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec("CREATE TABLE retained (id INTEGER)")
			return err
		},
		Down: func(tx *sql.Tx) error {
			if _, err := tx.Exec("DROP TABLE retained"); err != nil {
				return err
			}
			return errors.New("down failed")
		},
	}
	m.AddMigration(mg)
	if err := m.Up(0); err != nil {
		t.Fatal(err)
	}

	if err := m.Down(1); err == nil {
		t.Fatal("Down succeeded, want error")
	}
	if !mg.done {
		t.Fatal("failed rollback changed done state")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", count)
	}
	if _, err := db.Exec("SELECT * FROM retained"); err != nil {
		t.Fatalf("migration changes were not rolled back: %v", err)
	}
}

// TestDownStepCountsMigrationsNotBatches is the case the old test could not
// see, and the reason it could not: it applied its two migrations in two Up
// calls, so each batch held exactly one migration and "one batch" and "one
// migration" were the same number.
//
// Applied together, they are not. Down(1) used to take the latest *batch* —
// both of them — while the flag that reaches it says "Number of migrations".
// On a project whose schema went up in one `migrate up`, which is every
// project on its first deploy, asking to undo one migration dropped
// everything.
func TestDownStepCountsMigrationsNotBatches(t *testing.T) {
	db, m := newSQLiteMigrator(t)

	for _, version := range []string{"001", "002", "003"} {
		m.AddMigration(&Migration{
			Version: version,
			Up:      func(*sql.Tx) error { return nil },
			Down:    func(*sql.Tx) error { return nil },
		})
	}
	if _, err := Init(db, DriverSQLite); err != nil {
		t.Fatal(err)
	}

	// One batch, three migrations.
	if err := m.Up(0); err != nil {
		t.Fatal(err)
	}
	if got := countMigrations(t, db); got != 3 {
		t.Fatalf("applied = %d, want 3", got)
	}
	if got := countBatches(t, db); got != 1 {
		t.Fatalf("batches = %d, want 1 — the premise of this test", got)
	}

	if err := m.Down(1); err != nil {
		t.Fatal(err)
	}
	if got := countMigrations(t, db); got != 2 {
		t.Fatalf("after Down(1), applied = %d, want 2; a step is a migration, not a batch", got)
	}

	// And a step larger than one still stops where it was told to.
	if err := m.Down(1); err != nil {
		t.Fatal(err)
	}
	if got := countMigrations(t, db); got != 1 {
		t.Fatalf("after a second Down(1), applied = %d, want 1", got)
	}
}

// A bare Down takes the whole latest batch, which is the long-standing
// default and the one behaviour this change had to leave alone.
func TestDownWithNoStepTakesTheLatestBatch(t *testing.T) {
	db, m := newSQLiteMigrator(t)

	for _, version := range []string{"001", "002", "003"} {
		m.AddMigration(&Migration{
			Version: version,
			Up:      func(*sql.Tx) error { return nil },
			Down:    func(*sql.Tx) error { return nil },
		})
	}
	if _, err := Init(db, DriverSQLite); err != nil {
		t.Fatal(err)
	}

	// Two batches: one migration, then two.
	if err := m.Up(1); err != nil {
		t.Fatal(err)
	}
	if err := m.Up(0); err != nil {
		t.Fatal(err)
	}
	if got := countBatches(t, db); got != 2 {
		t.Fatalf("batches = %d, want 2", got)
	}

	if err := m.Down(0); err != nil {
		t.Fatal(err)
	}
	if got := countMigrations(t, db); got != 1 {
		t.Fatalf("after a bare Down, applied = %d, want 1 — the second batch held two", got)
	}
}

func countMigrations(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func countBatches(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(DISTINCT batch) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

// Registering one version twice registers it once.
//
// Migrations is a map and was always overwritten correctly; Versions is the
// ORDERED list the runner walks, and the version used to be appended to it
// unconditionally. So a second call ran that migration twice, and the second
// run met the table the first had just created:
//
//	Running migration 20261118160000
//	Finished running migration 20261118160000
//	Running migration 20261118160000
//	pq: relation "shifts" already exists
//
// Nothing caught it because every caller registered from init, which runs
// once per process. It appeared the moment a host started collecting
// migrations from compiled-in feature modules and calling AddMigration per
// composed application — which a test binary does several times.
func TestAddingOneVersionTwiceRegistersItOnce(t *testing.T) {
	m := &Migrator{Migrations: map[string]*Migration{}}

	first := &Migration{Version: "20260101000000"}
	second := &Migration{Version: "20260102000000"}

	m.AddMigration(first)
	m.AddMigration(second)
	m.AddMigration(first)

	if len(m.Versions) != 2 {
		t.Fatalf("Versions = %v, want two entries — the runner walks this list, so "+
			"a repeated version runs its migration twice", m.Versions)
	}
	if len(m.Migrations) != 2 {
		t.Errorf("Migrations holds %d, want 2", len(m.Migrations))
	}
	// And still sorted, which is what the insertion sort is for.
	if m.Versions[0] > m.Versions[1] {
		t.Errorf("Versions = %v, not in order", m.Versions)
	}
}

// Re-registering a version replaces what runs, rather than keeping the first
// registration.
//
// The map was always the authority, and this pins that: a host that composes
// an application twice gets the second composition's closures, which is what
// it asked for.
func TestReRegisteringAVersionReplacesIt(t *testing.T) {
	m := &Migrator{Migrations: map[string]*Migration{}}

	m.AddMigration(&Migration{Version: "20260101000000"})
	replacement := &Migration{Version: "20260101000000"}
	m.AddMigration(replacement)

	if m.Migrations["20260101000000"] != replacement {
		t.Error("the second registration did not replace the first")
	}
	if len(m.Versions) != 1 {
		t.Errorf("Versions = %v, want one entry", m.Versions)
	}
}

// Insertion order does not change the result, because the runner depends on
// Versions being sorted and a host hands modules over in whatever order its
// composition root lists them.
func TestVersionsAreSortedWhateverTheOrderOfRegistration(t *testing.T) {
	for _, order := range [][]string{
		{"3", "1", "2"},
		{"1", "2", "3"},
		{"2", "3", "1"},
		{"3", "2", "1"},
	} {
		m := &Migrator{Migrations: map[string]*Migration{}}
		for _, version := range order {
			m.AddMigration(&Migration{Version: version})
		}
		if got := strings.Join(m.Versions, ""); got != "123" {
			t.Errorf("registering %v gave %v", order, m.Versions)
		}
	}
}
