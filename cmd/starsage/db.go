package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// dbCmd represents the base command for database operations.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the local database.",
}

// resetCmd represents the command to reset the database.
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete and reset the local database file.",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting user home directory: %v\n", err)
			return
		}
		dbPath := filepath.Join(home, ".config", "starsage", "stars.db")

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Println("Database file does not exist. Nothing to do.")
			return
		}

		fmt.Printf("Are you sure you want to delete the database file at %s? [y/N]: ", dbPath)
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			err := os.Remove(dbPath)
			if err != nil {
				fmt.Printf("Error deleting database file: %v\n", err)
				return
			}
			fmt.Println("Database file successfully deleted.")
		} else {
			fmt.Println("Reset cancelled.")
		}
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(resetCmd)
}
