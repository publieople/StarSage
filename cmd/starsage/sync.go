package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync GitHub Stars to local database.",
	Long:  `Fetches all starred repositories from GitHub and saves them to a local SQLite database.
It supports incremental syncs to fetch only the new stars.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called. (GitHub API fetching and DB writing logic to be implemented here)")
		// TODO: Implement GitHub API fetching
		// TODO: Implement SQLite database writing
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
