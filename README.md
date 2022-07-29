# Go Migration

A simple to use database schema migration tool for go applications.

## Installation

Open your favorite terminal app and cd into your go module enabled project root:

`cd packagename`

[Replace "packagename" with your package's name]

To install the migration package, run:

`go get -u github.com/lemmego/migration`

## Usage

This package requires DSN (Data Source Name) and database driver name in order to function properly. There are two ways you can provide these. Either via `DATABASE_URL` and `DB_DRIVER` environment variables:

```env
DATABASE_URL=username:password@protocol(address)/dbname?param=value
DB_DRIVER=mysql
```

Or, via command line flags:

```sh
go run . migrate up --dsn="username:password@protocol(address)/dbname?param=value" --driver=mysql
```

_[A Note Aboute DSN: The DSN should be compatible with the database driver you wish to use. Refer to the corresponding driver's documentation to see what's the accepted format for your particular db driver. In the examples below, the "https://github.com/go-sql-driver/mysql" driver will be used and their format will be followed.]_

In all the examples below, we will assume that the values are coming from the environment variables.

In your go mod enabled application, import the db driver, godotenv (if you prefer env vars over flags) and the "github.com/lemmego/migration/cmd" package. Then add the `cmd.Execute()` statement.

```go
// main.go

package main

import (
	"log"

	// _ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	// _ "packagename/migrations"

	"github.com/joho/godotenv"
	"github.com/lemmego/migration/cmd"
)

func main() {
	err := godotenv.Load()
  	if err != nil {
    		log.Fatal("Error loading .env file")
  	}
	cmd.Execute()
}
```

Then run `go run . migrate create -n create_users_table` to create a sample migration. A migration file will be created inside the `migrations` directory of the project root. If the `migrations` directory is not present in your project root, the `migrate create ...` command will create one for you. Open the migration file and populate the `up()` and `down()` method like this:

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

Now open the `main.go` file and uncomment (or add, if it's missing) the following blank import statement:

```go
// _ "packagename/migrations"
```

This blank import will make sure that migration files residing inside the `migrations/` directory will execute. Replace "packagename" with your package name (can be found in the topmost line your go.mod file).

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

## Credits

This package is an enhancement of this blog post by Praveen:
https://techinscribed.com/create-db-migrations-tool-in-go-from-scratch/

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
