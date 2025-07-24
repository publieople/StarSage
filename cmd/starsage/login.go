package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"star-sage/internal/config"
	"star-sage/internal/gh"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub.",
	Long:  `Authenticate with GitHub using OAuth 2.0 Device Flow to get an access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting GitHub authentication...")
		token, err := gh.PerformDeviceFlow(context.Background(), proxyURL)
		if err != nil {
			fmt.Printf("Error during authentication: %v\n", err)
			return
		}

		if err := config.SaveToken(token); err != nil {
			fmt.Printf("Error saving token: %v\n", err)
			return
		}

		fmt.Println("Successfully authenticated and token saved.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
