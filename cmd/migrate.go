package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/lemmego/migration"
	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "DB Schema Migration Tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new empty migrations file",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Unable to read flag `name`", err.Error())
			return
		}
		if err := migration.Create(name); err != nil {
			fmt.Println("Unable to create migration", err.Error())
			return
		}
	},
}

func GetDsnAndDriver(cmd *cobra.Command) (string, string) {
	var dsnStr, driver string

	dsnStr, err := cmd.Flags().GetString("dsn")
	if err != nil {
		log.Fatalf("Unable to read flag `dsn`", err.Error())
	}

	if dsnStr == "" {
		dsnStr = os.Getenv("DATABASE_URL")
	}

	driver, err = cmd.Flags().GetString("driver")
	if err != nil {
		log.Fatalf("Unable to read flag `driver`", err.Error())
	}
	if driver == "" {
		if val, found := os.LookupEnv("DB_DRIVER"); found {
			driver = val
		} else {
			log.Fatalf("Driver is required. Either pass it as flag (--driver) or set the DB_DRIVER environment variable.")
		}
	}

	if dsnStr == "" {
		dsn := migration.NewDsnBuilder(driver).
			SetHost(os.Getenv("DB_HOST")).
			SetPort(os.Getenv("DB_PORT")).
			SetUsername(os.Getenv("DB_USERNAME")).
			SetPassword(os.Getenv("DB_PASSWORD")).
			SetName(os.Getenv("DB_DATABASE")).
			Build()

		switch driver {
		case "mysql":
			dsnStr = dsn.GetMysqlDSN()
		case "postgres":
			dsnStr = dsn.GetPostgresDSN()
			log.Println("[DSN:", dsnStr, "]")
		case "sqlite3":
			dsnStr = dsn.GetSqliteDSN()
		default:
			log.Fatalf("Driver not supported")
		}
	}

	return dsnStr, driver
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "run up migrations",
	Run: func(cmd *cobra.Command, args []string) {
		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`", err.Error())
			return
		}
		dsnStr, driver := GetDsnAndDriver(cmd)
		db := migration.NewDB(dsnStr, driver)
		if db == nil {
			fmt.Println("Unable to connect to the database")
			return
		}

		migrator, err := migration.Init(db, driver)
		if err != nil {
			fmt.Println("Unable to fetch migrator", err.Error())
			return
		}

		err = migrator.Up(step)
		if err != nil {
			fmt.Println("Unable to run `up` migrations", err.Error())
			return
		}
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "run down migrations",
	Run: func(cmd *cobra.Command, args []string) {
		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`", err.Error())
			return
		}

		dsnStr, driver := GetDsnAndDriver(cmd)
		db := migration.NewDB(dsnStr, driver)

		migrator, err := migration.Init(db, driver)
		if err != nil {
			fmt.Println("Unable to fetch migrator", err.Error())
			return
		}

		err = migrator.Down(step)
		if err != nil {
			fmt.Println("Unable to run `down` migrations", err.Error())
			return
		}
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "display status of each migrations",
	Run: func(cmd *cobra.Command, args []string) {
		dsn, driver := GetDsnAndDriver(cmd)
		db := migration.NewDB(dsn, driver)

		migrator, err := migration.Init(db, driver)
		if err != nil {
			fmt.Println("Unable to fetch migrator", err.Error())
			return
		}

		if err := migrator.MigrationStatus(); err != nil {
			fmt.Println("Unable to fetch migration status", err.Error())
			return
		}
	},
}

func init() {
	// Add "--name", "--driver" and "--dsn" flags to "create" command
	migrateCreateCmd.Flags().StringP("name", "n", "", "Name for the migration")
	migrateCreateCmd.Flags().StringP("driver", "d", "", "Driver Name")
	migrateCreateCmd.Flags().StringP("dsn", "u", "", "Data Source Name")

	// Add "--step", "--driver" and "--dsn" flags to "up" and "down" command
	migrateUpCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")
	migrateUpCmd.Flags().StringP("driver", "d", "", "Data Source Name")
	migrateUpCmd.Flags().StringP("dsn", "u", "", "Data Source Name")

	migrateDownCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")
	migrateDownCmd.Flags().StringP("driver", "d", "", "Driver Name")
	migrateDownCmd.Flags().StringP("dsn", "u", "", "Data Source Name")

	// Add "--driver" and "--dsn" flags to "status" command
	migrateStatusCmd.Flags().StringP("driver", "d", "", "Driver Name")
	migrateStatusCmd.Flags().StringP("dsn", "u", "", "Data Source Name")

	// Add "create", "status", "up" and "down" commands to the "migrate" command
	rootCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateCreateCmd, migrateStatusCmd)
}
