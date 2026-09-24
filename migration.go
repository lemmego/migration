package migration

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

//go:embed template.txt
var stub string

// Migration represents a migration data type
type Migration struct {
	Version string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error

	done bool
}

// Migrator is a struct that holds the migrations
type Migrator struct {
	mu         sync.Mutex
	db         *sql.DB
	dialect    string
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
	m.mu.Lock()
	defer m.mu.Unlock()

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

	m := migrator
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db = db
	m.dialect = dialect

	// Create `schema_migrations` table to remember which migrations were executed.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version varchar(255),
		batch int
	);`); err != nil {
		fmt.Println("Unable to create `schema_migrations` table", err)
		return m, err
	}

	// Find out all the executed migrations
	rows, err := db.Query("SELECT version FROM schema_migrations;")
	if err != nil {
		return m, err
	}

	defer rows.Close()

	// Mark the migrations as Done if it is already executed
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return m, err
		}

		if m.Migrations[version] != nil {
			m.Migrations[version].done = true
		}
	}

	if err := rows.Err(); err != nil {
		return m, err
	}

	return m, nil
}

// Up method runs the migrations which have not yet been run
func (m *Migrator) Up(step int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var bindPlaceHolders string
	if m.dialect == DriverMySQL || m.dialect == DriverSQLite {
		bindPlaceHolders = "?, ?"
	} else if m.dialect == DriverPostgres {
		bindPlaceHolders = "$1, $2"
	} else {
		return errors.New("unsupported driver")
	}

	// Use background context for transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	count := 0
	lastBatch := 0
	if rows, err := tx.Query("SELECT MAX(batch) FROM schema_migrations;"); err != nil {
		return err
	} else {
		for rows.Next() {
			var lastBatchPtr *int // use a pointer to int to allow for NULL values
			if err := rows.Scan(&lastBatchPtr); err != nil {
				return err
			}
			if lastBatchPtr != nil {
				lastBatch = *lastBatchPtr // dereference the pointer to get the actual value
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
	}

	var completed []*Migration
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
		completed = append(completed, mg)
		fmt.Println("Finished running migration", mg.Version)

		count++
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	for _, mg := range completed {
		mg.done = true
	}

	return nil
}

// Down migration rolls back the last batch of migrations
func (m *Migrator) Down(step int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var bindPlaceHolder string
	switch m.dialect {
	case DriverMySQL, DriverSQLite:
		bindPlaceHolder = "?"
	case DriverPostgres:
		bindPlaceHolder = "$1"
	default:
		return errors.New("unsupported driver")
	}

	// Use background context for transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// A step is a number of batches, with zero preserving the default of the
	// latest batch. The lower bound is inclusive, so step=1 selects one batch.
	batchOffset := step
	if batchOffset < 1 {
		batchOffset = 1
	}
	rows, err := tx.Query(fmt.Sprintf(
		`SELECT version FROM schema_migrations WHERE batch >= (SELECT MAX(batch) - %s + 1 FROM schema_migrations) AND batch <= (SELECT MAX(batch) FROM schema_migrations) ORDER BY batch DESC, version DESC;`,
		bindPlaceHolder), batchOffset)
	if err != nil {
		return err
	}

	var reverted []*Migration
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
		reverted = append(reverted, mg)
		fmt.Println("Finished reverting migration", mg.Version)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, mg := range reverted {
		mg.done = false
	}

	return nil
}

// Status checks which migrations have run and which have not
func (m *Migrator) MigrationStatus() error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
