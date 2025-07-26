package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"star-sage/internal/db"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search your starred repositories.",
	Long: `Performs a full-text search on the name, description, and README
of your starred repositories stored in the local database.`,
	Args: cobra.MinimumNArgs(1), // Require at least one argument for the query
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		fmt.Printf("Searching for: \"%s\"\n\n", query)

		database, err := db.InitDB()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer database.Close()

		results, err := db.SearchRepositories(database, query, limit)
		if err != nil {
			fmt.Printf("Error performing search: %v\n", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		fmt.Printf("Found %d results:\n", len(results))
		for _, repo := range results {
			fmt.Printf("----------------------------------------\n")
			fmt.Printf("Repo: %s\n", repo.FullName)
			fmt.Printf("URL: %s\n", repo.URL)
			if repo.Description != "" {
				fmt.Printf("Description: %s\n", repo.Description)
			}
			if repo.Summary != "" {
				fmt.Printf("AI Summary: %s\n", repo.Summary)
			}
		}
		fmt.Printf("----------------------------------------\n")
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
