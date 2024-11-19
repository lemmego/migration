package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "DB Schema Migration Tool",
	Version: "0.1.7",
}

// Execute ..
func Execute() {
	if err := MigrateCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}
