# Go Migration

A simple to use database schema migration tool for go applications.

## Installation

Open your favorite terminal app and cd into your go module enabled project root:

`cd packagename`

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
```

The supported `DB_DRIVER` values are `sqlite`, `mysql` and `postgres`

```go
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

Then run `go run ./cmd/migratations create create_users_table` to create a sample migration. A migration file will be created inside the `projectroot/cmd/migrations/` directory. If the `migrations` directory is not present in your project root, the `migrate create ...` command will create one for you. Open the migration file and populate the `up()` and `down()` method like this:

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

Optionally, you could also use the db agnostic schema builder API:

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
    t.Integer("org_id")
    t.String("first_name", 255)
    t.String("last_name", 255)
    t.String("email", 255).Unique()
    t.String("password", 255)
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

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
