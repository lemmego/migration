package cmd

import (
	"fmt"
	"os"

	"github.com/lemmego/migration"
	"github.com/spf13/cobra"
)

var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new empty migrations file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Migration name is required")
			return
		}
		name := args[0]
		// name, err := cmd.Flags().GetString("name")
		// if err != nil {
		// 	fmt.Println("Unable to read flag `name`", err.Error())
		// 	return
		// }
		if err := migration.CreateMigration(name); err != nil {
			fmt.Println("Unable to create migration", err.Error())
			return
		}
	},
}

func GetDriver(cmd *cobra.Command) (string, error) {
	driver, err := cmd.Flags().GetString("driver")
	if err != nil {
		return "", err
	}

	if driver == "" {
		driver = os.Getenv("DB_DRIVER")
	}

	if driver == "" {
		return "", fmt.Errorf("driver is required")
	}

	return driver, nil
}

func GetDSN(cmd *cobra.Command, driver string) (string, error) {
	dsnStr, err := cmd.Flags().GetString("dsn")
	if err != nil {
		return "", err
	}

	if dsnStr == "" {
		dsnStr = os.Getenv("DATABASE_URL")
	}

	if dsnStr == "" {
		ds := migration.DataSource{
			Driver:   driver,
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_DATABASE"),
			Params:   os.Getenv("DB_PARAMS"),
		}
		dsnStr, err = ds.String()
		if err != nil {
			return "", err
		}
	}

	return dsnStr, nil
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

		driver, err := GetDriver(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		dsnStr, err := GetDSN(cmd, driver)
		if err != nil {
			fmt.Println(err)
			return
		}

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

		driver, err := GetDriver(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		dsnStr, err := GetDSN(cmd, driver)
		if err != nil {
			fmt.Println(err)
			return
		}

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
		driver, err := GetDriver(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		dsnStr, err := GetDSN(cmd, driver)
		if err != nil {
			fmt.Println(err)
			return
		}

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
	MigrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateCreateCmd, migrateStatusCmd)
}
