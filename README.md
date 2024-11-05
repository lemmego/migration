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
