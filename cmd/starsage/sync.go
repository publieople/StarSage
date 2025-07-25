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

		fmt.Printf("Found %d starred repositories. Fetching READMEs and saving to database...\n", len(repos))

		database, err := db.InitDB()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer database.Close()

		// We need a client to fetch READMEs
		client, err := gh.NewClient(proxyURL, token)
		if err != nil {
			fmt.Printf("Error creating GitHub client: %v\n", err)
			return
		}

		for i, repo := range repos {
			fmt.Printf("[%d/%d] Syncing %s...\n", i+1, len(repos), repo.FullName)

			readmeContent, err := gh.GetReadme(context.Background(), client, repo.FullName)
			if err != nil {
				fmt.Printf("Could not get README for %s: %v. Skipping.\n", repo.FullName, err)
				// still save the repo metadata even if README fails
			}

			dbRepo := db.Repository{
				ID:              repo.ID,
				FullName:        repo.FullName,
				Description:     repo.Description,
				URL:             repo.HTMLURL,
				Language:        repo.Language,
				StargazersCount: repo.StargazersCount,
				ReadmeContent:   readmeContent,
			}

			if err := db.UpsertRepository(database, dbRepo); err != nil {
				fmt.Printf("Error saving repository %s to database: %v\n", repo.FullName, err)
				// Decide if we should continue or stop. For now, let's continue.
			}
		}

		fmt.Println("Successfully synced and saved repositories to the local database.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
