package migration

import (
	"database/sql"
	"fmt"
	"os"
)

func getDsnFromEnv() string {
	if os.Getenv("DB_DRIVER") == "mysql" {
		return fmt.Sprintf("%s:%s@/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	} else if os.Getenv("DB_DRIVER") == "postgres" {
		return fmt.Sprintf(
			"%s://%s:%s@%s:%s/%s",
			os.Getenv("DB_DRIVER"),
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME")+"?sslmode=disable",
		)
	} else {
		return fmt.Sprintf("%s:%s@/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	}
}

func NewDB() *sql.DB {
	fmt.Println("Connecting to database...")

	db, err := sql.Open(os.Getenv("DB_DRIVER"), getDsnFromEnv())
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
