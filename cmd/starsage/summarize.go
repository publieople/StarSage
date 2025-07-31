package main

import (
	"context"
	"fmt"
	"net/http"
	"star-sage/internal/ai"
	"star-sage/internal/db"

	"github.com/spf13/cobra"
)

var (
	aiProvider string
	aiModel    string
)

// summarizeCmd represents the summarize command
var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Summarize starred repositories using an AI provider.",
	Long: `Reads repository data (like READMEs) from the local database,
sends it to a specified AI provider to generate a summary,
and saves the summary back to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting summarization process...")

		database, err := db.InitDB()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer database.Close()

		summarizeLimit := limit
		if summarizeLimit == 0 {
			summarizeLimit = 5 // Default to 5 if no limit is set
		}
		fmt.Printf("Attempting to summarize up to %d repositories...\n", summarizeLimit)

		repos, err := db.GetReposForSummarization(database, summarizeLimit)
		if err != nil {
			fmt.Printf("Error getting repositories to summarize: %v\n", err)
			return
		}

		if len(repos) == 0 {
			fmt.Println("No new repositories to summarize.")
			return
		}

		fmt.Printf("Found %d repositories to summarize.\n", len(repos))

		// The AI provider might need its own http client (without auth or proxy)
		aiClient := &http.Client{}
		var provider ai.Provider
		switch aiProvider {
		case "ollama":
			provider = ai.NewOllamaProvider(aiModel, "", aiClient) // Use default Ollama URL
		default:
			fmt.Printf("Unsupported AI provider: %s\n", aiProvider)
			return
		}

		for _, repo := range repos {
			fmt.Printf("Summarizing %s...\n", repo.FullName)
			// We need the full README content, but our GetReposForSummarization only gets a subset of fields.
			// This is a flaw in the current logic. For the MVP, we'll assume ReadmeContent is populated.
			// A proper implementation would fetch the full repo details here if needed.
			if repo.ReadmeContent == "" {
				fmt.Printf("Skipping %s, no README content found in DB.\n", repo.FullName)
				continue
			}

			// Build the prompt manually now that the provider is generic
			prompt := fmt.Sprintf("Please provide a concise summary of the following project's README, focusing on its purpose and key features. Output only the summary text:\n\n---\n\n%s", repo.ReadmeContent)

			summary, err := provider.Generate(context.Background(), prompt)
			if err != nil {
				fmt.Printf("Error summarizing %s: %v\n", repo.FullName, err)
				continue // Move to the next repo
			}

			if err := db.UpdateRepoSummary(database, repo.ID, summary); err != nil {
				fmt.Printf("Error saving summary for %s: %v\n", repo.FullName, err)
				continue
			}
			fmt.Printf("Successfully summarized and saved for %s.\n\n", repo.FullName)
		}
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
	summarizeCmd.Flags().StringVar(&aiProvider, "provider", "ollama", "The AI provider to use (e.g., ollama, openai)")
	summarizeCmd.Flags().StringVar(&aiModel, "model", "llama3:8b", "The specific model to use for summarization")
}
