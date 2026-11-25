package cmd

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/lemmego/migration"
	"github.com/spf13/cobra"
)

// dsnCommand builds a command carrying the --dsn flag GetDSN reads, so these
// tests exercise the real resolution order rather than a copy of it.
func dsnCommand(dsn string) *cobra.Command {
	cmd := &cobra.Command{Use: "migrate"}
	cmd.Flags().String("dsn", dsn, "")
	return cmd
}

const (
	unpooled = "postgres://u:p@ep-example.eu-central-1.aws.neon.tech/db?sslmode=require"
	pooled   = "postgres://u:p@ep-example-pooler.eu-central-1.aws.neon.tech/db?sslmode=require"
)

// The whole point of the change. Neon and Supabase publish a pooled URL and a
// direct one, and their documentation steers applications at the pooled one —
// so DATABASE_URL in a real deployment is usually the endpoint a migration
// must not use. Preferring the unpooled variable makes the common deployment
// correct without the operator having to know any of that.
func TestGetDSNPrefersTheUnpooledURL(t *testing.T) {
	t.Setenv("DATABASE_URL", pooled)
	t.Setenv("DATABASE_URL_UNPOOLED", unpooled)

	got, err := GetDSN(dsnCommand(""), "postgres")
	if err != nil {
		t.Fatalf("GetDSN() error = %v", err)
	}
	if got != unpooled {
		t.Errorf("GetDSN() = %q, want the unpooled URL %q", got, unpooled)
	}
}

// Most projects set only DATABASE_URL, and that has to keep working exactly as
// it did. A change that required a second variable would break every existing
// deployment.
func TestGetDSNFallsBackToDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", pooled)
	t.Setenv("DATABASE_URL_UNPOOLED", "")

	got, err := GetDSN(dsnCommand(""), "postgres")
	if err != nil {
		t.Fatalf("GetDSN() error = %v", err)
	}
	if got != pooled {
		t.Errorf("GetDSN() = %q, want %q", got, pooled)
	}
}

// Somebody naming an endpoint has said which one they mean, including when
// they deliberately want the pooled one.
func TestGetDSNFlagBeatsBothVariables(t *testing.T) {
	t.Setenv("DATABASE_URL", pooled)
	t.Setenv("DATABASE_URL_UNPOOLED", unpooled)

	const explicit = "postgres://u:p@127.0.0.1:5432/other?sslmode=disable"
	got, err := GetDSN(dsnCommand(explicit), "postgres")
	if err != nil {
		t.Fatalf("GetDSN() error = %v", err)
	}
	if got != explicit {
		t.Errorf("GetDSN() = %q, want the flag value %q", got, explicit)
	}
}

// With no URL at all the DB_* components still build one, which is how a
// scaffolded project with a local Postgres works.
func TestGetDSNFallsBackToComponents(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_UNPOOLED", "")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USERNAME", "someone")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_DATABASE", "appdb")
	t.Setenv("DB_PARAMS", "sslmode=disable")

	got, err := GetDSN(dsnCommand(""), "postgres")
	if err != nil {
		t.Fatalf("GetDSN() error = %v", err)
	}
	if got == "" {
		t.Fatal("GetDSN() returned an empty DSN")
	}
	for _, want := range []string{"127.0.0.1", "appdb", "someone"} {
		if !strings.Contains(got, want) {
			t.Errorf("GetDSN() = %q, missing %q", got, want)
		}
	}
}

// An empty DATABASE_URL_UNPOOLED must not shadow DATABASE_URL. A present but
// empty variable is a stray line in a .env, not a decision — the same trap
// config.MustEnv has fallen into twice in this repo.
func TestGetDSNTreatsAnEmptyUnpooledURLAsAbsent(t *testing.T) {
	t.Setenv("DATABASE_URL_UNPOOLED", "")
	t.Setenv("DATABASE_URL", pooled)

	got, err := GetDSN(dsnCommand(""), "postgres")
	if err != nil {
		t.Fatalf("GetDSN() error = %v", err)
	}
	if got != pooled {
		t.Errorf("GetDSN() = %q, want %q", got, pooled)
	}
}

func TestMigrateFreshCommandIsRegisteredWithDatabaseFlags(t *testing.T) {
	var fresh *cobra.Command
	for _, command := range MigrateCmd.Commands() {
		if command.Name() == "fresh" {
			fresh = command
			break
		}
	}
	if fresh == nil {
		t.Fatal("migrate fresh command is not registered")
	}
	if fresh.Flags().Lookup("driver") == nil || fresh.Flags().Lookup("dsn") == nil {
		t.Fatal("migrate fresh must accept the same driver and DSN options as up/down")
	}
	if fresh.Flags().Lookup("path") == nil {
		t.Fatal("migrate fresh must accept --path")
	}
}

func TestMigrationLifecycleCommandsExposePathFlag(t *testing.T) {
	for _, name := range []string{"up", "down", "fresh", "status", "create"} {
		t.Run(name, func(t *testing.T) {
			var command *cobra.Command
			for _, candidate := range MigrateCmd.Commands() {
				if candidate.Name() == name {
					command = candidate
					break
				}
			}
			if command == nil || command.Flags().Lookup("path") == nil {
				t.Fatalf("migrate %s must expose --path", name)
			}
		})
	}
}

func TestSelectMigrationPathsConfiguresMigrator(t *testing.T) {
	migrator := migration.GetMigrator()
	version := "path-selection-test"
	if err := migrator.AddMigrationIn("cli_test", &migration.Migration{Version: version, Up: func(*sql.Tx) error { return nil }, Down: func(*sql.Tx) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "up"}
	cmd.Flags().StringSlice("path", []string{"cli_test"}, "")
	if err := selectMigrationPaths(cmd, migrator); err != nil {
		t.Fatalf("selectMigrationPaths() error = %v", err)
	}
	if err := migrator.UsePaths(); err != nil {
		t.Fatalf("reset migration selection: %v", err)
	}
}

func TestMigrationCommandsUseSelectedPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "cli-migrations.db")
	migrator := migration.GetMigrator()
	for _, migrationSpec := range []struct {
		path, version, table string
	}{
		{path: "cli_lifecycle_platform", version: "cli-platform-001", table: "cli_platform_table"},
		{path: "cli_lifecycle_tenant", version: "cli-tenant-001", table: "cli_tenant_table"},
	} {
		spec := migrationSpec
		if err := migrator.AddMigrationIn(spec.path, &migration.Migration{
			Version: spec.version,
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec("CREATE TABLE " + spec.table + " (value TEXT)")
				return err
			},
			Down: func(tx *sql.Tx) error {
				_, err := tx.Exec("DROP TABLE " + spec.table)
				return err
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = migrator.UsePaths() })

	configure := func(cmd *cobra.Command) {
		t.Helper()
		for flag, value := range map[string]string{"driver": "sqlite", "dsn": dbPath, "path": "cli_lifecycle_tenant"} {
			if err := cmd.Flags().Set(flag, value); err != nil {
				t.Fatalf("set %s: %v", flag, err)
			}
		}
	}
	configure(migrateUpCmd)
	migrateUpCmd.Run(migrateUpCmd, nil)

	assertTable := func(table string, wantPresent bool) {
		t.Helper()
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		var name string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name)
		if wantPresent && err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
		if !wantPresent && err == nil {
			t.Fatalf("table %s unexpectedly exists", table)
		}
	}
	assertTable("cli_tenant_table", true)
	assertTable("cli_platform_table", false)

	configure(migrateFreshCmd)
	if err := migrateFreshCmd.RunE(migrateFreshCmd, nil); err != nil {
		t.Fatalf("fresh selected path: %v", err)
	}
	assertTable("cli_tenant_table", true)
	assertTable("cli_platform_table", false)

	configure(migrateDownCmd)
	migrateDownCmd.Run(migrateDownCmd, nil)
	assertTable("cli_tenant_table", false)
	assertTable("cli_platform_table", false)

}
