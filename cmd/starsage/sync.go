package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"star-sage/internal/config"
	"star-sage/internal/db"
	"star-sage/internal/gh"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync GitHub Stars to local database.",
	Long:  `Fetches all starred repositories from GitHub and saves them to a local SQLite database.
It supports incremental syncs to fetch only the new stars.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := config.GetToken()
		if token == "" {
			fmt.Println("Authentication token not found. Please run 'starsage login' first.")
			return
		}

		fmt.Println("Syncing GitHub Stars...")
		repos, err := gh.GetStarredRepos(context.Background(), token, proxyURL)
		if err != nil {
			fmt.Printf("Error syncing stars: %v\n", err)
			return
		}

		fmt.Printf("Found %d starred repositories. Saving to database...\n", len(repos))

		database, err := db.InitDB()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer database.Close()

		if err := db.SaveRepositories(database, repos); err != nil {
			fmt.Printf("Error saving repositories to database: %v\n", err)
			return
		}

		fmt.Println("Successfully synced and saved repositories to the local database.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
