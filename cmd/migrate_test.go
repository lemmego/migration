package cmd

import (
	"strings"
	"testing"

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
