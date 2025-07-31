package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"star-sage/internal/ai"
	"star-sage/internal/db"
	"strconv"
	"strings"
)

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, the response is likely already partially sent.
		// Log the error, but don't try to send another http.Error.
		fmt.Printf("Error encoding JSON response: %v\n", err)
	}
}

// writeError is a helper to write JSON error responses.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// apiHandler creates a http.HandlerFunc that shares a database connection.
type apiHandler struct {
	db *sql.DB
}

// StartServer starts the web server on the given port.
func StartServer(port int) error {
	database, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	// The database connection is closed when the application exits.
	// defer database.Close() is not used here as it would close immediately.

	h := &apiHandler{db: database}
	mux := http.NewServeMux()

	// API handlers
	mux.HandleFunc("/api/repositories", h.handleGetRepositories)
	mux.HandleFunc("/api/lists", h.handleLists) // Will handle GET (all) and POST
	mux.HandleFunc("/api/lists/", h.handleListByID) // Will handle GET (by ID)

	// Static file server
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func (h *apiHandler) handleGetRepositories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	repos, err := db.GetAllRepositories(h.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error fetching repositories")
		return
	}
	writeJSON(w, http.StatusOK, repos)
}

func (h *apiHandler) handleLists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetLists(w, r)
	case http.MethodPost:
		h.handleCreateList(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *apiHandler) handleGetLists(w http.ResponseWriter, r *http.Request) {
	lists, err := db.GetLists(h.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error fetching lists")
		return
	}
	writeJSON(w, http.StatusOK, lists)
}

type createListRequest struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

func (h *apiHandler) handleCreateList(w http.ResponseWriter, r *http.Request) {
	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "List name and prompt are required")
		return
	}

	// For now, we'll do this synchronously. Asynchronous tasks can be added later.
	fmt.Printf("Received request to create list '%s' with prompt: %s\n", req.Name, req.Prompt)

	listID, err := db.CreateList(h.db, req.Name, req.Prompt)
	if err != nil {
		// This could be a UNIQUE constraint violation, handle it gracefully.
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "A list with this name already exists")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to create list in database")
		}
		return
	}

	fmt.Printf("List '%s' created with ID: %d. Starting classification...\n", req.Name, listID)
	// In a real app, this would be a background job.
	// For simplicity, we run it in a goroutine and don't wait.
	go h.runClassification(listID, req.Prompt)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "List creation initiated. Classification is running in the background.",
		"list_id": listID,
	})
}

// runClassification is a helper to perform the AI classification in the background.
func (h *apiHandler) runClassification(listID int64, prompt string) {
	// This function runs in a goroutine, so it needs its own error handling.
	fmt.Printf("Fetching all repositories for classification...\n")
	repos, err := db.GetAllRepositories(h.db)
	if err != nil {
		fmt.Printf("[Error][List %d] Failed to get repositories: %v\n", listID, err)
		return
	}

	// Note: In a real app, provider/model would come from config or the request.
	// Here we hardcode it for simplicity.
	provider := ai.NewOllamaProvider("llama3:8b", "", &http.Client{})

	fmt.Printf("[List %d] Classifying %d repositories...\n", listID, len(repos))
	classifiedRepoIDs, err := ai.ClassifyRepositories(context.Background(), provider, prompt, repos)
	if err != nil {
		fmt.Printf("[Error][List %d] AI classification failed: %v\n", listID, err)
		return
	}

	fmt.Printf("[List %d] Found %d matching repositories. Saving to database...\n", listID, len(classifiedRepoIDs))
	if len(classifiedRepoIDs) > 0 {
		if err := db.AddReposToList(h.db, listID, classifiedRepoIDs); err != nil {
			fmt.Printf("[Error][List %d] Failed to add repos to list: %v\n", listID, err)
			return
		}
	}

	fmt.Printf("[List %d] Classification completed successfully.\n", listID)
}

func (h *apiHandler) handleListByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/lists/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid list ID")
		return
	}

	repos, err := db.GetReposByListID(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error fetching repositories for the list")
		return
	}
	writeJSON(w, http.StatusOK, repos)
}
