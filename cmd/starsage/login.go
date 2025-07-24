package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
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

		fmt.Printf("Successfully authenticated! Token: %s\n", token)
		// TODO: Save the token to a config file
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
