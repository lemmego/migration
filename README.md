# Migration

## Installation

Open your favorite terminal app and cd into your go module enabled project root:

`cd myproject`

To install the migration package, run:

`go get -u github.com/lemmego/migration`

## Usage

The following environment variables are needed in order for the migration to work:

```.env
DB_DRIVER=mysql
DB_USERNAME=root
DB_PASSWORD=
DB_HOST=localhost
DB_PORT=3306
DB_NAME=testdb
```

```go
package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
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
