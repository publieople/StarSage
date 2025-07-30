package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"star-sage/internal/db"
)

// StartServer starts the web server on the given port.
func StartServer(port int) error {
	mux := http.NewServeMux()

	// API handlers
	mux.HandleFunc("/api/repositories", handleGetRepositories)

	// Static file server
	// This will be configured later to serve from an embedded filesystem or a 'frontend' directory.
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func handleGetRepositories(w http.ResponseWriter, r *http.Request) {
	database, err := db.InitDB()
	if err != nil {
		http.Error(w, "Error initializing database", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	repos, err := db.GetAllRepositories(database)
	if err != nil {
		http.Error(w, "Error fetching repositories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
