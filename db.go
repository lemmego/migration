package migration

import (
	"database/sql"
	"fmt"
	"log"
)

func NewDB(dsn string, driverName string) *sql.DB {
	fmt.Println("Connecting to database...")
	fmt.Println("DSN:", dsn)

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err.Error())
	}

	// if err := db.Ping(); err != nil {
	// 	fmt.Printf("%+v", db)
	// 	log.Fatal("Unable to ping the database: ", err.Error())
	// }

	fmt.Println("Database connected!")

	return db
}
