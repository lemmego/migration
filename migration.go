package migration

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed template.txt
var stub string

var dbDialect string

// Migration represents a migration data type
type Migration struct {
	Version string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error

	done bool
}

// Migrator is a struct that holds the migrations
type Migrator struct {
	db         *sql.DB
	Versions   []string
	Migrations map[string]*Migration
}

var migrator = &Migrator{
	Versions:   []string{},
	Migrations: map[string]*Migration{},
}

// GetMigrator returns the migrator
func GetMigrator() *Migrator {
	return migrator
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(mg *Migration) {
	// Add the migration to the hash with version as key
	m.Migrations[mg.Version] = mg

	// Insert version into versions array using insertion sort
	index := 0
	for index < len(m.Versions) {
		if m.Versions[index] > mg.Version {
			break
		}
		index++
	}

	m.Versions = append(m.Versions, mg.Version)
	copy(m.Versions[index+1:], m.Versions[index:])
	m.Versions[index] = mg.Version
}

// Init populates the fields of Migrator and returns it
func Init(db *sql.DB, dialect string) (*Migrator, error) {
	if dialect != DriverSQLite && dialect != DriverMySQL && dialect != DriverPostgres {
		return nil, errors.New("unsupported driver")
	}

	dbDialect = dialect
	migrator.db = db

	// Create `schema_migrations` table to remember which migrations were executed.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version varchar(255),
		batch int
	);`); err != nil {
		fmt.Println("Unable to create `schema_migrations` table", err)
		return migrator, err
	}

	// Find out all the executed migrations
	rows, err := db.Query("SELECT version FROM schema_migrations;")
	if err != nil {
		return migrator, err
	}

	defer rows.Close()

	// Mark the migrations as Done if it is already executed
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return migrator, err
		}

		if migrator.Migrations[version] != nil {
			migrator.Migrations[version].done = true
		}
	}

	return migrator, err
}

// Up method runs the migrations which have not yet been run
func (m *Migrator) Up(step int) error {
	var bindPlaceHolders string
	if dbDialect == DriverMySQL || dbDialect == DriverSQLite {
		bindPlaceHolders = "?, ?"
	} else if dbDialect == DriverPostgres {
		bindPlaceHolders = "$1, $2"
	} else {
		return errors.New("unsupported driver")
	}

	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	count := 0
	lastBatch := 0
	if rows, err := m.db.Query("SELECT MAX(batch) FROM schema_migrations;"); err != nil {
		return err
	} else {
		defer rows.Close()
		for rows.Next() {
			var lastBatchPtr *int // use a pointer to int to allow for NULL values
			if err := rows.Scan(&lastBatchPtr); err != nil {
				return err
			}
			if lastBatchPtr != nil {
				lastBatch = *lastBatchPtr // dereference the pointer to get the actual value
			}
		}
	}

	for _, v := range m.Versions {
		if step > 0 && count == step {
			break
		}

		mg := m.Migrations[v]
		if mg.done {
			continue
		}

		fmt.Println("Running migration", mg.Version)
		if err := mg.Up(tx); err != nil {
			tx.Rollback()
			return err
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations VALUES("+bindPlaceHolders+")", mg.Version, lastBatch+1); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Println("Finished running migration", mg.Version)

		count++
	}

	tx.Commit()

	return nil
}

// Down migration rolls back the last batch of migrations
func (m *Migrator) Down(step int) error {
	var bindPlaceHolder string
	if dbDialect == DriverMySQL || dbDialect == DriverSQLite {
		bindPlaceHolder = "?"
	} else if dbDialect == DriverPostgres {
		bindPlaceHolder = "$1"
	} else {
		return errors.New("unsupported driver")
	}
	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	// Reverse the migration based on the batch column and the step passed

	rows, err := m.db.Query(
		fmt.Sprintf(`SELECT version FROM schema_migrations WHERE batch BETWEEN (SELECT MAX(batch - %s) FROM schema_migrations) AND (SELECT MAX(batch) FROM schema_migrations) ORDER BY version DESC;`, bindPlaceHolder),
		step,
	)
	if err != nil {
		return err
	}

	defer rows.Close()

	var version string
	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return err
		}

		mg := m.Migrations[version]
		if !mg.done {
			return errors.New("migration not found")
		}

		fmt.Println("Reverting Migration", mg.Version)
		if err := mg.Down(tx); err != nil {
			tx.Rollback()
			return err
		}

		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = "+bindPlaceHolder, mg.Version); err != nil {
			tx.Rollback()
			return err
		}
		fmt.Println("Finished reverting migration", mg.Version)
	}

	tx.Commit()

	return nil
}

// Status checks which migrations have run and which have not
func (m *Migrator) MigrationStatus() error {
	for _, v := range m.Versions {
		mg := m.Migrations[v]

		if mg.done {
			fmt.Println(fmt.Sprintf("Migration %s... completed", v))
		} else {
			fmt.Println(fmt.Sprintf("Migration %s... pending", v))
		}
	}

	return nil
}

// guessPackageNameFromMigrationsDir guesses the package name from a given migrations dir path.
func guessPackageNameFromMigrationsDir(migrationsDir string) string {
	splitPath := strings.Split(migrationsDir, "/")
	return splitPath[len(splitPath)-1]
}

// CreateMigration creates a migration file
func CreateMigration(name string) error {
	migrationsDir := os.Getenv("MIGRATIONS_DIR")

	if migrationsDir != "" {
		migrationsDir = strings.TrimSuffix(migrationsDir, "/")
	} else {
		migrationsDir = "./cmd/migrations"
	}

	packageName := guessPackageNameFromMigrationsDir(migrationsDir)

	version := time.Now().Format("20060102150405")

	in := struct {
		Version     string
		Name        string
		PackageName string
	}{
		Version:     version,
		Name:        name,
		PackageName: packageName,
	}

	var out bytes.Buffer
	tx := template.New("template")
	t := template.Must(tx.Parse(stub))
	err := t.Execute(&out, in)
	if err != nil {
		return errors.New("Unable to execute template:" + err.Error())
	}
	wd, _ := os.Getwd()
	path := filepath.Join(wd, migrationsDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return errors.New("Unable to create migrations directory:" + err.Error())
		}
	}
	f, err := os.Create(fmt.Sprintf("%s/%s_%s.go", path, version, name))
	if err != nil {
		return errors.New("Unable to create migration file:" + err.Error())
	}
	defer f.Close()

	if _, err := f.WriteString(out.String()); err != nil {
		return errors.New("Unable to write to migration file:" + err.Error())
	}

	fmt.Println("Generated new migration files...", f.Name())
	return nil
}

func reverse(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
