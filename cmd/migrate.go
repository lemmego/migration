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

	// DATABASE_URL_UNPOOLED first, because a migration must not run through a
	// transaction pooler.
	//
	// Neon and Supabase both publish two URLs for one database: a pooled one
	// fronted by PgBouncer in transaction mode, and a direct one. Their
	// documentation steers applications at the pooled URL, so that is what
	// lands in a deployment's environment as DATABASE_URL — and it is the
	// wrong endpoint for this command twice over.
	//
	// lib/pq issues every parameterised query on the *unnamed* prepared
	// statement over the extended protocol. Under transaction pooling a Parse
	// from one logical connection pairs with a Bind from another, which fails
	// as "bind message supplies N parameters, but prepared statement \"\"
	// requires M" — intermittently, and in both directions. The migrator's own
	// bookkeeping queries are parameterised, so a schema change can fail
	// halfway through a deploy and report a message about a statement nobody
	// wrote.
	//
	// Pooled connections are also session-less by design, so anything a
	// migration does that outlives a statement — an advisory lock, a SET, a
	// session-scoped temp table — silently applies to whichever backend
	// answered rather than to the migration.
	//
	// An explicit --dsn still wins, because somebody naming an endpoint has
	// said which one they mean.
	if dsnStr == "" {
		dsnStr = os.Getenv("DATABASE_URL_UNPOOLED")
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
