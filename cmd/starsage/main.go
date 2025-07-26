package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"star-sage/internal/config"
)

var (
	proxyURL string
	limit    int
)

var rootCmd = &cobra.Command{
	Use:   "starsage",
	Short: "StarSage is a tool to manage your GitHub Stars.",
	Long: `A Fast and Flexible CLI for managing, searching, and summarizing your GitHub Stars.
Complete documentation is available at https://github.com/publieople/StarSage`, // Placeholder URL
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.InitConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is given
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&proxyURL, "proxy", "", "HTTP proxy to use for network requests (e.g. http://127.0.0.1:7890)")
	rootCmd.PersistentFlags().IntVar(&limit, "limit", 0, "Limit the number of items to process (0 for no limit)")
}

func main() {
	Execute()
}
