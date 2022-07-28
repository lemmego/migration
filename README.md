# Migration

## Installation

Step 1:

Open your favorite terminal app and cd into your go module enabled project root:

`cd myproject`

To install the migration package, run:

`go get -u github.com/lemmego/migration`

Step 2:

Next, create a subpackage at location `myproject/cmd/migration/

`mkdir cmd/migration && touch cmd/migration/migrate.go`

Step 3:

Open the newly create migrate.go file and paste the following content:

`// cmd/migration/migrate.go`

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/lemmego/migration"

	_ "github.com/go-sql-driver/mysql"
)

func create(name string) {
	if err := migration.Create(name); err != nil {
		fmt.Println("Unable to create migration", err.Error())
		return
	}
}

func up(step int) {
	db := migration.NewDB()

	migrator, err := migration.Init(db)
	if err != nil {
		fmt.Println("Unable to fetch migrator", err.Error())
		return
	}

	err = migrator.Up(step)
	if err != nil {
		fmt.Println("Unable to run `up` migrations", err.Error())
		return
	}
}
func down(step int) {
	db := migration.NewDB()

	migrator, err := migration.Init(db)
	if err != nil {
		fmt.Println("Unable to fetch migrator", err.Error())
		return
	}

	err = migrator.Down(step)
	if err != nil {
		fmt.Println("Unable to run `down` migrations", err.Error())
		return
	}
}

func main() {
	err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {
		switch argsWithoutProg[0] {
		case "create":
			if len(argsWithoutProg) < 2 {
				fmt.Println("Unable to create migration, no name provided")
				return
			} else {
				create(argsWithoutProg[1])
			}
		case "up":
			if len(argsWithoutProg) < 2 {
				up(1)
			} else {
				step, err := strconv.Atoi(argsWithoutProg[1])
				if err != nil {
					fmt.Println("Step must be a number", err.Error())
					return
				}
				up(step)
			}
		case "down":
			if len(argsWithoutProg) < 2 {
				down(1)
			} else {
				step, err := strconv.Atoi(argsWithoutProg[1])
				if err != nil {
					fmt.Println("Step must be a number", err.Error())
					return
				}
				down(step)
			}
		}
	}
}
```

Step 4:

To create a new migration file, run:

`go run cmd/migration/* create create_users_table`

A new migration file with timestamp will be created inside the cmd/migration directory with `up()` and `down()` methods. The `up()` method runs when the new db schema changes are need to be pushed and the `down()` runs when the last run migrations need to be rolled back. Populate the `up()` and the `down()` methods. Example:

```go
package main

import (
	"database/sql"

	"github.com/lemmego/migration"
)

func init() {
	migration.GetMigrator().AddMigration(&migration.Migration{
		Version: "20220728233614",
		Up:      mig_20220728233614_create_users_table_up,
		Down:    mig_20220728233614_create_users_table_down,
	})
}

func mig_20220728233614_create_users_table_up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE users ( name varchar(255) );")
	if err != nil {
		return err
	}
	return nil
}

func mig_20220728233614_create_users_table_down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE users")
	if err != nil {
		return err
	}
	return nil
}
```

Step 5:

Run this migration with:

`go run cmd/migration/* up`

Open the database, and you should see the users table with the name column.

If you want to rollback:

`go run cmd/migration/* down`

Refresh the database, and the users table will be dropped.
