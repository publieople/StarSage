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
		repos, err := gh.GetStarredRepos(context.Background(), token, proxyURL, limit)
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

		// Pre-fetch existing etags to avoid querying the DB in a loop
		existingRepos, err := db.GetAllReposWithETags(database)
		if err != nil {
			fmt.Printf("Warning: could not pre-fetch existing repo data: %v\n", err)
		}
		etagMap := make(map[int64]string)
		readmeMap := make(map[int64]string)
		for _, r := range existingRepos {
			etagMap[r.ID] = r.ETag
			readmeMap[r.ID] = r.ReadmeContent
		}

		for i, repo := range repos {
			fmt.Printf("[%d/%d] Syncing %s...\n", i+1, len(repos), repo.FullName)

			currentEtag := etagMap[repo.ID]
			readmeContent, newEtag, err := gh.GetReadme(context.Background(), client, repo.FullName, currentEtag)
			if err != nil {
				fmt.Printf("Could not get README for %s: %v. Skipping.\n", repo.FullName, err)
			}

			dbRepo := db.Repository{
				ID:              repo.ID,
				FullName:        repo.FullName,
				Description:     repo.Description,
				URL:             repo.HTMLURL,
				Language:        repo.Language,
				StargazersCount: repo.StargazersCount,
				ETag:            newEtag,
			}

			// If README was not modified, use the old content from the map.
			if newEtag == currentEtag && currentEtag != "" {
				dbRepo.ReadmeContent = readmeMap[repo.ID]
			} else {
				dbRepo.ReadmeContent = readmeContent
			}

			if err := db.UpsertRepository(database, dbRepo); err != nil {
				fmt.Printf("Error saving repository %s to database: %v\n", repo.FullName, err)
			}
		}

		fmt.Println("Successfully synced and saved repositories to the local database.")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
