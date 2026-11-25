package migration

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"go/token"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed template.txt
var stub string

// Migration represents a migration data type
type Migration struct {
	Version string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error

	done bool
	path string
}

// Migrator is a struct that holds the migrations
type Migrator struct {
	mu            sync.Mutex
	db            *sql.DB
	dialect       string
	Versions      []string
	Migrations    map[string]*Migration
	selectedPaths map[string]struct{}
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
// Dialect reports the database the migrator was initialised for.
//
// A migration that writes its own DDL needs this: the schema builder reads
// DB_DRIVER, but a package shipping a migration renders through
// SchemaStatements or CreateFor and has to be told. Empty before Init runs.
func (m *Migrator) Dialect() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dialect
}

// AddMigration registers a migration, replacing any already registered under
// the same version.
//
// Idempotent, which it was not: Migrations is a map keyed on the version and
// was overwritten correctly, but Versions is the ORDERED list the runner
// walks and the version was appended to it unconditionally. So a second call
// for one version ran that migration twice — and the second run met the
// table the first had just created.
//
// It went unnoticed because every caller in this ecosystem registered from
// init, which runs once per process. It surfaced the moment a host began
// collecting migrations from compiled-in feature modules and calling this
// per composed application, which a test binary does several times. A
// registration function that duplicates on a second call is a trap for
// exactly that pattern.
func (m *Migrator) AddMigration(mg *Migration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mg.path = ""
	m.addMigrationLocked(mg)
}

// AddMigrationIn registers a migration in a named, relative migration set.
// Migration versions are global to the compiled registry, including across
// paths, because schema_migrations stores versions without path metadata.
func (m *Migrator) AddMigrationIn(pathName string, mg *Migration) error {
	normalized, err := normalizeMigrationPath(pathName)
	if err != nil {
		return err
	}
	if mg == nil {
		return errors.New("migration: cannot register a nil migration")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing := m.Migrations[mg.Version]; existing != nil && existing.path != normalized {
		return fmt.Errorf("migration: version %q is already registered in path %q", mg.Version, existing.path)
	}
	mg.path = normalized
	m.addMigrationLocked(mg)
	return nil
}

// UsePaths selects the registered migration sets used by Up, Down, Fresh,
// and MigrationStatus. Calling UsePaths with no arguments restores the legacy
// behavior of selecting every registered migration.
func (m *Migrator) UsePaths(paths ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(paths) == 0 {
		m.selectedPaths = nil
		return nil
	}

	selected := make(map[string]struct{}, len(paths))
	for _, pathName := range paths {
		normalized, err := normalizeMigrationPath(pathName)
		if err != nil {
			return err
		}
		found := false
		for _, mg := range m.Migrations {
			if mg.path == normalized {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("migration: no migrations registered in path %q", normalized)
		}
		selected[normalized] = struct{}{}
	}
	m.selectedPaths = selected
	return nil
}

func normalizeMigrationPath(pathName string) (string, error) {
	if pathName == "" || strings.Contains(pathName, "\\") || strings.HasPrefix(pathName, "/") || path.IsAbs(pathName) {
		return "", fmt.Errorf("migration: path %q must be a non-empty relative slash-separated path", pathName)
	}
	cleaned := path.Clean(pathName)
	if cleaned == "." || cleaned != pathName || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("migration: invalid path %q", pathName)
	}
	for _, segment := range strings.Split(cleaned, "/") {
		if !token.IsIdentifier(segment) {
			return "", fmt.Errorf("migration: path segment %q must be a valid Go identifier", segment)
		}
	}
	return cleaned, nil
}

func (m *Migrator) addMigrationLocked(mg *Migration) {
	_, already := m.Migrations[mg.Version]
	m.Migrations[mg.Version] = mg
	if already {
		return
	}
	index := 0
	for index < len(m.Versions) && m.Versions[index] <= mg.Version {
		index++
	}
	m.Versions = append(m.Versions, mg.Version)
	copy(m.Versions[index+1:], m.Versions[index:])
	m.Versions[index] = mg.Version
}

func (m *Migrator) selectedMigrationsLocked() []*Migration {
	selected := make([]*Migration, 0, len(m.Versions))
	for _, version := range m.Versions {
		mg := m.Migrations[version]
		if mg == nil {
			continue
		}
		if m.selectedPaths != nil {
			if _, ok := m.selectedPaths[mg.path]; !ok {
				continue
			}
		}
		selected = append(selected, mg)
	}
	return selected
}

// Init populates the fields of Migrator and returns it
func Init(db *sql.DB, dialect string) (*Migrator, error) {
	// Normalised, so that Init(db, "libsql") or Init(db, "pgx") works rather
	// than being refused for a spelling. The error names the value it was
	// given, which "unsupported driver" alone did not.
	normalised, ok := NormaliseDialect(dialect)
	if !ok {
		return nil, fmt.Errorf("migration: %w %q, want one of %s",
			ErrUnsupportedDialect, dialect, strings.Join(SupportedDialects(), ", "))
	}
	dialect = normalised

	m := migrator
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db = db
	m.dialect = dialect
	m.selectedPaths = nil

	// Create `schema_migrations` table to remember which migrations were executed.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version varchar(255),
		batch int
	);`); err != nil {
		fmt.Println("Unable to create `schema_migrations` table", err)
		return m, err
	}

	// The database is the authority on what has run, so start from "nothing
	// has" and let it say otherwise.
	//
	// Without the reset, done only ever moved from false to true, and the
	// migrator is a package-level singleton. So a process that initialised
	// against a second database — a test harness with a scratch database per
	// case, or a program migrating one tenant after another — carried the
	// first database's done flags into the second and silently skipped every
	// migration, leaving an empty schema and no error.
	for _, mg := range m.Migrations {
		mg.done = false
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
	for _, mg := range m.selectedMigrationsLocked() {
		if step > 0 && count == step {
			break
		}

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

// Down reverts migrations, most recently applied first.
//
// step is a number of migrations, as it is for Up. Zero reverts the whole of
// the latest batch, which is what a bare `migrate down` means.
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

	// A step is a number of *migrations*, counted back from the most recently
	// applied, and zero means the whole of the latest batch.
	//
	// Counting batches instead would be the same thing only when every batch
	// holds one migration. On a project whose schema went up in a single
	// `migrate up` — every project on its first deploy — the latest batch is
	// the entire schema, so `--step 1` would revert all of it. Up counts
	// migrations too, and the flag reaching both says "Number of migrations
	// to execute".
	//
	// The test that was supposed to cover this applied its two migrations in
	// two separate Up calls, so one batch held one migration and both
	// readings agreed. A single-batch case is the one that tells them apart.
	rows, err := tx.Query(`SELECT version, batch FROM schema_migrations ORDER BY batch DESC, version DESC;`)
	if err != nil {
		return err
	}

	// Drain the cursor completely before running anything else on this
	// transaction.
	//
	// A *sql.Tx holds one connection, and a statement issued while a Rows
	// from the same Tx is still open desynchronises the wire protocol. With
	// lib/pq that surfaces as "unexpected Parse response 'C'" on the first
	// statement the reverted migration runs, so `migrate down` failed on
	// Postgres every time while passing on MySQL and SQLite, whose drivers
	// buffer the result set and hide the overlap.
	selectedPaths := m.selectedPaths
	var applied []struct {
		version string
		batch   int
	}
	for rows.Next() {
		var item struct {
			version string
			batch   int
		}
		if err := rows.Scan(&item.version, &item.batch); err != nil {
			return err
		}
		mg := m.Migrations[item.version]
		if mg == nil {
			if selectedPaths != nil {
				continue
			}
			applied = append(applied, item)
			continue
		}
		if selectedPaths != nil {
			if _, ok := selectedPaths[mg.path]; !ok {
				continue
			}
		}
		applied = append(applied, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	var versions []string
	if step > 0 {
		if step < len(applied) {
			applied = applied[:step]
		}
	} else if len(applied) > 0 {
		latestBatch := applied[0].batch
		for _, item := range applied {
			if item.batch != latestBatch {
				break
			}
			versions = append(versions, item.version)
		}
	}
	if step > 0 {
		for _, item := range applied {
			versions = append(versions, item.version)
		}
	}

	var reverted []*Migration
	for _, version := range versions {
		mg := m.Migrations[version]
		if mg == nil || !mg.done {
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

	if err := tx.Commit(); err != nil {
		return err
	}
	for _, mg := range reverted {
		mg.done = false
	}

	return nil
}

// Fresh reverses every registered migration currently applied to the database,
// then applies the complete current migration set. It is intentionally
// destructive: application data is removed when each migration's Down runs.
// Init must be called first so the database's applied migrations are known.
func (m *Migrator) Fresh() error {
	if m == nil || m.db == nil {
		return errors.New("migration: Fresh requires Init to be called first")
	}
	m.mu.Lock()
	selected := m.selectedMigrationsLocked()
	var applied int
	for _, mg := range selected {
		if mg.done {
			applied++
		}
	}
	m.mu.Unlock()
	if applied > 0 {
		if err := m.Down(applied); err != nil {
			return fmt.Errorf("migration: reverting applied migrations for fresh: %w", err)
		}
	}
	if err := m.Up(0); err != nil {
		return fmt.Errorf("migration: applying migrations for fresh: %w", err)
	}
	return nil
}

// Status checks which migrations have run and which have not
func (m *Migrator) MigrationStatus() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, mg := range m.selectedMigrationsLocked() {
		v := mg.Version

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
	return filepath.Base(filepath.Clean(migrationsDir))
}

// CreateMigration creates a migration file
func CreateMigration(name string) error {
	return createMigration("", name)
}

// CreateMigrationIn creates a migration in the configured migration root's
// relative path and emits an AddMigrationIn registration.
func CreateMigrationIn(pathName, name string) error {
	normalized, err := normalizeMigrationPath(pathName)
	if err != nil {
		return err
	}
	return createMigration(normalized, name)
}

func createMigration(pathName, name string) error {
	if !token.IsIdentifier(name) {
		return fmt.Errorf("migration: invalid migration name %q", name)
	}
	migrationsDir := os.Getenv("MIGRATIONS_DIR")

	if migrationsDir != "" {
		migrationsDir = strings.TrimSuffix(migrationsDir, "/")
	} else {
		migrationsDir = "./cmd/migrations"
	}

	packageName := guessPackageNameFromMigrationsDir(migrationsDir)
	if pathName != "" {
		packageName = filepath.Base(pathName)
	}

	version := time.Now().Format("20060102150405")

	in := struct {
		Version     string
		Name        string
		PackageName string
		Path        string
	}{
		Version:     version,
		Name:        name,
		PackageName: packageName,
		Path:        pathName,
	}

	var out bytes.Buffer
	tx := template.New("template")
	t := template.Must(tx.Parse(stub))
	err := t.Execute(&out, in)
	if err != nil {
		return errors.New("Unable to execute template:" + err.Error())
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("migration: get working directory: %w", err)
	}
	baseDir := migrationsDir
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(wd, baseDir)
	}
	root, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("migration: resolve migrations root: %w", err)
	}
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		return fmt.Errorf("migration: create migrations root: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("migration: resolve migrations root: %w", err)
	}
	resolvedOutput := resolvedRoot
	if pathName != "" {
		for _, segment := range strings.Split(pathName, "/") {
			resolvedOutput = filepath.Join(resolvedOutput, segment)
			if err := os.Mkdir(resolvedOutput, os.ModePerm); err != nil && !os.IsExist(err) {
				return fmt.Errorf("migration: create migrations directory: %w", err)
			}
			info, err := os.Lstat(resolvedOutput)
			if err != nil {
				return fmt.Errorf("migration: inspect migrations directory: %w", err)
			}
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("migration: path component %q is not a real directory", segment)
			}
		}
	}
	f, err := os.Create(filepath.Join(resolvedOutput, fmt.Sprintf("%s_%s.go", version, name)))
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
