# Go Migration

A simple to use database schema migration tool for go applications.

## Installation

Open your favorite terminal app and cd into your go module enabled project root:

`cd projectroot`

[Replace "packagename" with your package's name]

To install the migration package, run:

`go get -u github.com/lemmego/migration`

## Usage

This package resolves the DSN (Data Source Name) from the following env variables:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=test
DB_USERNAME=root
DB_PASSWORD=
DB_PARAMS=charset=utf8mb4&collation=utf8mb4_unicode_ci
MIGRATIONS_DIR="./cmd/migrations" # Optional
```

The supported `DB_DRIVER` values are `sqlite`, `mysql` and `postgres`

```go
// projectroot/main.go
// Or,
// projectroot/cmd/myapp/main.go

package main

import (
  "github.com/joho/godotenv"
  "github.com/lemmego/migration/cmd"
  _ "projectroot/cmd/migrations"
)

func main() {
  if err := godotenv.Load(); err != nil {
    panic(err)
  }
  cmd.Execute()
}
```

Next, to create a sample migration, run:

`go run . create create_users_table` (If the main.go is located in projectroot/main.go)

Or,

`go run ./cmd/myapp create create_users_table` (If the main.go is located in projectroot/cmd/myapp/main.go)

**Note:** For all the examples (_migrate up, migrate down, migrate status_) below, we will assume that the `main.go` is located in `projectroot/main.go` and the `go run . <subcommand>` command will be used. Replace this with `go run ./cmd/myapp <subcommand>` if your main.go is in `projectroot/cmd/myapp/main.go`

If you didn't provide a `MIGRATIONS_DIR` env variable, A migration file will be created inside the `projectroot/cmd/migrations/` directory. If the directory is not present in your project root, the `migrate create ...` command will create one for you. Open the migration file and populate the `up()` and `down()` method like this:

```go
// 20220729200658_create_users_table.go

package migrations

import (
	"database/sql"
	"github.com/lemmego/migration"
)

func init() {
	migration.GetMigrator().AddMigration(&migration.Migration{
		Version: "20220729200658",
		Up:      mig_20220729200658_create_users_table_up,
		Down:    mig_20220729200658_create_users_table_down,
	})
}

func mig_20220729200658_create_users_table_up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE users ( name varchar(255) );")
	if err != nil {
		return err
	}
	return nil
}

func mig_20220729200658_create_users_table_down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE users")
	if err != nil {
		return err
	}
	return nil
}
```

Alternatively, you could also use the db agnostic schema builder API. Which is useful for switching databases without having to rewrite the migrations.

```go
// 20220729200658_create_users_table.go

package migrations

import (
	"database/sql"
	"github.com/lemmego/migration"
)

func init() {
	migration.GetMigrator().AddMigration(&migration.Migration{
		Version: "20220729200658",
		Up:      mig_20220729200658_create_users_table_up,
		Down:    mig_20220729200658_create_users_table_down,
	})
}

func mig_20220729200658_create_users_table_up(tx *sql.Tx) error {
  schema := migration.Create("users", func(t *migration.Table) {
    t.BigIncrements("id").Primary()
    t.ForeignID("org_id").Constrained() // "org_id" references the "id" column in the "orgs" table
    t.String("first_name", 255)
    t.String("last_name", 255)
    t.String("email", 255).Unique()
    t.String("password", 255)
    t.Text("bio").Nullable()
    t.DateTime("created_at", 0).Default("now()")
    t.DateTime("updated_at", 0).Default("now()")
  }).Build()

  if _, err := tx.Exec(schema); err != nil {
    return err
  }

  return nil
}

func mig_20220729200658_create_users_table_down(tx *sql.Tx) error {
  schema := migration.Drop("users").Build()
  if _, err := tx.Exec(schema); err != nil {
    return err
  }
  return nil
}
```

Once you've made sure that the expected environment variables are present in your `.env` file, you can run `go run . migrate up`

You should see something like the following:

```
Connecting to database...
Database connected!
Running migration 20220729200658
Finished running migration 20220729200658
```

Open your database client application (e.g. SequelPro, TablePlus) and open the database. You should see two new tables: schema_migrations and users.

You can revert the migration by running `go run . migrate down`. You should see something like this:

```
Connecting to database...
Database connected!
Reverting Migration 20220729200658
Finished reverting migration 20220729200658
```

Both the `migrate up` and the `migrate down` commands take a `--step` integer flag to indicate how many step should the migration run forward or backward:

E.g.:

`go run . migrate down --step=1`

There is also a `migrate status` command to see which migrations are currently pending and/or completed.

### Indexes, keys and constraints

Everything in this section renders per dialect, which is the reason to say it
here rather than in raw SQL: the same declaration becomes a separate
`CREATE INDEX` on Postgres and SQLite and an inline `INDEX` clause on MySQL,
and the same integer becomes `BIGSERIAL` or `BIGINT UNSIGNED` or `BIGINT`. A
hand-written statement is correct for exactly one vendor.

```go
migration.Create("endorsements", func(t *migration.Table) {
    t.UnsignedBigInt("endorser_id")
    t.UnsignedBigInt("user_id")
    t.UnsignedBigInt("skill_id")
    t.String("note", 140).Default("")
    t.TimestampTz("created_at", 0).Default(migration.Expr("CURRENT_TIMESTAMP"))

    // A value set, a range, a format. CHECK is standard SQL: Postgres,
    // MySQL 8.0.16+ and SQLite all enforce one. Enum() does not — it emits
    // no constraint at all on Postgres — so this is how a closed set is
    // expressed.
    t.Check("endorsements_not_self", "endorser_id <> user_id")

    // A document column. Postgres gets JSONB, MySQL gets JSON, SQLite gets
    // TEXT — which is not a fallback but the way SQLite stores JSON, since
    // its JSON1 functions operate on TEXT.
    t.Json("payload").Nullable()

    // A composite foreign key: one constraint over the pair, which two
    // single-column keys do not express. Neither of them would refuse a row
    // naming a user and a skill that are never paired.
    //
    // Name is opt-in. Without it the key is anonymous and the server derives
    // one, which is what every key this package emitted before v0.5.0 did —
    // so naming by default would rename all of them. Use it to restate a
    // constraint that already exists under a name of its own.
    t.Foreign("user_id", "skill_id").
        Name("endorsements_claimed_skill").
        References("user_id", "skill_id").
        On("user_skills").
        OnDelete("CASCADE")

    // An index, with the name it already has in somebody's database. The
    // derived name is right for a new table and wrong for an existing one.
    t.Index("user_id", "created_at").Name("endorsements_recent_idx")

    // Newest first. Postgres, MySQL 8.0 and SQLite 3.8.3 all honour a
    // descending key. Desc names a column already in the index — writing the
    // direction into the column list works only because the builder joins
    // its arguments verbatim, and breaks the day it quotes an identifier.
    t.Index("user_id", "created_at").
        Desc("created_at").
        Name("endorsements_feed_idx")

    // A unique *index*, which is not a unique constraint. On Postgres they
    // are different catalog objects, so a schema being matched rather than
    // merely created has to produce the right one; MySQL implements the
    // constraint with an index, which is what makes the distinction easy to
    // lose. Use UniqueKey when the uniqueness is part of what the table is,
    // and this when it is a lookup index that happens to be unique.
    t.Index("skill_id", "user_id").Unique().Name("endorsements_pair_key")

    // The constraint form.
    t.UniqueKey("user_id", "note").Name("endorsements_one_note")

    // A partial index, covering only the rows that match. Postgres and
    // SQLite have them; MySQL has none, and Validate refuses one there
    // rather than dropping the predicate — an index silently covering every
    // row gives correct results and a plan nobody can account for.
    t.Index("user_id").
        Where("note <> ''").
        Name("endorsements_annotated_idx")
})
```

`Validate` refuses two declarations that cannot mean what they say on any
dialect, rather than letting the driver report them from further away: a
`Desc` naming a column the index does not carry, and a composite foreign key
referencing a different number of columns than it carries. `Build` calls
`Validate` and panics; call `Validate` yourself to check without panicking.

### Altering a table

```go
migration.Alter("posts", func(t *migration.Table) {
    t.Int("view_count").Default(0)              // ADD COLUMN
    t.String("slug", 160).Nullable().Change()   // change type, nullability, default
    t.DropColumn("legacy_body")
    t.RenameColumn("body", "body_source")

    t.DropIndex("posts_author_idx")
    t.DropUniqueKey("posts_slug_unique")
    t.DropForeignKey("posts_author_fk")
    t.DropCheck("posts_kind_valid").IfExists()  // for a down migration
    t.DropPrimaryKey()
})
```

Each of these is spelled differently by the three dialects, which is most of
the reason to declare rather than write them:

| | Postgres | MySQL | SQLite |
|---|---|---|---|
| drop a primary key | `DROP CONSTRAINT <table>_pkey` | `DROP PRIMARY KEY` | refused |
| drop a unique key | `DROP CONSTRAINT <name>` | `DROP INDEX <name>` | refused |
| drop a foreign key | `DROP CONSTRAINT <name>` | `DROP FOREIGN KEY <name>` | refused |
| drop an index | `DROP INDEX <name>`, its own statement | `ALTER TABLE … DROP INDEX <name>` | `DROP INDEX <name>` |
| change a column | `ALTER COLUMN … TYPE …`, plus `SET`/`DROP NOT NULL` | `MODIFY COLUMN …` | refused |
| several actions | one statement, comma-separated | one statement, comma-separated | **one statement each** |

`IfExists` is for a down migration, which should not fail because the thing it
removes is already gone. MySQL has no `IF EXISTS` on any of these — not on an
`ALTER TABLE` drop action and not on `DROP INDEX` — so `Validate` refuses it
there rather than quietly dropping the tolerance.

`AlterFor(dialect, table, fn)` builds one for a named dialect, the way
`CreateFor` does, which is how the table above is tested.

#### What still needs raw SQL

Migrations are `func(tx *sql.Tx) error`, so `tx.Exec` is the sanctioned
escape hatch. These have no portable form and are not pretending to:

- **Expression and operator-class indexes**, `tsvector`, `GIN`/`GiST` —
  Postgres only.
- **Functions and triggers.** `plpgsql` is Postgres, MySQL's trigger syntax
  differs, and SQLite has no stored functions at all.
- **Generated columns**, whose expression is dialect-specific.
- **`COMMENT ON`**. Postgres and MySQL have it in different shapes; SQLite
  has none.
- **Altering a column or a constraint on SQLite**, which means rebuilding the
  table. Refused rather than rendered, because what SQLite's `ALTER TABLE`
  accepts is `RENAME TO`, `RENAME COLUMN`, `ADD COLUMN` and `DROP COLUMN`,
  and nothing else.

#### A note on auto-increment keys

`BigIncrements` and `Increments` emit the dialect's own auto-increment type
and no extra constraint. Before v0.4.0 the Postgres path also appended
`CHECK (id > 0)`, so one declaration produced three different schemas — a
CHECK on Postgres, an `UNSIGNED` column type on MySQL, nothing on SQLite —
and a generated table could not be matched by a hand-written one. A sequence
starting at 1 and ascending cannot violate it, so nothing is lost.

`Column.Unsigned()` on Postgres emits nothing, and never has:
`NewDataType` appends suffixes before the column is marked, the same ordering
that makes `Enum` silent. MySQL honours it, because its `UNSIGNED` is read
when the type is rendered.

### Adding "migrate" command to an existing command:
If your project already has a command, say `rootCmd`, you could add the `MigrateCmd` to that command to take full control of the package:

```go
package mypackage

import "github.com/spf13/cobra"
import "github.com/lemmego/migration/cmd"

var rootCmd = &cobra.Command{}

func init() {
    rootCmd.AddCommand(cmd.MigrateCmd)
	rootCmd.Execute()
}
```

### Renaming package for self-contained binary:
By default, the migration files will be created within the `./cmd/migrations` directory. You can override the directory with the `MIGRATIONS_DIR` env variable. The package name of the migration files will follow the Go's convention of adopting the package name according to the directory the files are in, meaning if they are generated in a "migrations" directory, the package name will be "migrations". If you would like the package name to be `"main"` so that you can deploy it as a self-contained binary, follow these two steps:

1. Rename the package name of each migration files to `package main`
2. Add the following `main.go` file to the same directory where migrations are generated:

```go
// Assuming your generated files are in the projectroot/cmd/migrations dir:
// projectroot/cmd/migrations/main.go

package main

import (
	"github.com/joho/godotenv"
	"github.com/lemmego/migration/cmd"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	cmd.Execute()
}
```

Now the package `./cmd/migrations` is ready to be built and deployed as an independent binary.

Build:

`go build ./cmd/migrations`

Run:

`./migrations migrate create/up/down/status`



## Documentation

The package documentation can be found at: https://pkg.go.dev/github.com/lemmego/migration

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
