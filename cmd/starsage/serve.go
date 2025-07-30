package main

import (
	"fmt"
	"star-sage/internal/server"

	"github.com/spf13/cobra"
)

var port int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a web server to browse and manage your stars.",
	Long:  `Starts a local web server that provides a UI for viewing, searching, and managing your starred repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting server on port %d...\n", port)
		if err := server.StartServer(port); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}
