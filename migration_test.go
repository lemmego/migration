package migration

import (
	"database/sql"
	"errors"
	"path/filepath"
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
