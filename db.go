package migration

import (
	"database/sql"
	"fmt"
	"log"
)

// NewDB creates a new database connection.
//
// The driver must be registered, which means your application blank-imports
// it — this package does not, and that is deliberate. It used to import
// sqlite, MySQL and Postgres together, so every project linked all three and
// paid for the two it had never configured; the pure-Go sqlite driver alone
// is twenty-nine packages of translated libc.
//
//	import _ "github.com/lib/pq"            // or
//	import _ "github.com/glebarez/go-sqlite"
//	import _ "github.com/go-sql-driver/mysql"
//
// A missing driver fails here rather than later, on every environment at
// once, with database/sql's own message naming the forgotten import.
func NewDB(dsn string, driverName string) *sql.DB {
	fmt.Println("Connecting to database...")
	fmt.Println("DSN:", dsn)

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err.Error())
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Unable to ping the database: ", err.Error())
	}

	fmt.Println("Database connected!")

	return db
}
