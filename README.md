# Migration

## Installation

Open your favorite terminal app and cd into your go module enabled project root:

`cd myproject`

To install the migration package, run:

`go get -u github.com/lemmego/migration`

## Usage

```go
package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/lemmego/migration/cmd"
)

func main() {
	cmd.Execute()
}
```
