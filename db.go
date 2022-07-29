package migration

import (
	"database/sql"
	"fmt"
)

func NewDB(dsn string, driverName string) *sql.DB {
	fmt.Println("Connecting to database...")
	
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		fmt.Println("Unable to connect to database", err.Error())
		return nil
	}

	if err := db.Ping(); err != nil {
		fmt.Println("Unable to connect to database", err.Error())
		return nil
	}

	fmt.Println("Database connected!")

	return db
}
