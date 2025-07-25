package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const dbFileName = "stars.db"

// Repository represents a starred GitHub repository.
type Repository struct {
	ID              int64
	FullName        string
	Description     string
	URL             string
	Language        string
	StargazersCount int
	ReadmeContent   string
	Summary         string
	LastSyncedAt    string
}

// InitDB initializes the SQLite database and creates tables if they don't exist.
func InitDB() (*sql.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}
	dbPath := filepath.Join(home, ".config", "starsage", dbFileName)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	if err = createTables(db); err != nil {
		return nil, fmt.Errorf("could not create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	const createTableSQL = `
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER NOT NULL PRIMARY KEY,
		full_name TEXT NOT NULL UNIQUE,
		description TEXT,
		url TEXT,
		language TEXT,
		stargazers_count INTEGER,
		readme_content TEXT,
		summary TEXT,
		last_synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	return err
}

// UpsertRepository inserts or updates a single repository in the database.
func UpsertRepository(db *sql.DB, repo Repository) error {
	stmt, err := db.Prepare(`
		INSERT INTO repositories (id, full_name, description, url, language, stargazers_count, readme_content, last_synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			full_name=excluded.full_name,
			description=excluded.description,
			url=excluded.url,
			language=excluded.language,
			stargazers_count=excluded.stargazers_count,
			readme_content=excluded.readme_content,
			last_synced_at=excluded.last_synced_at;
	`)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		repo.ID,
		repo.FullName,
		repo.Description,
		repo.URL,
		repo.Language,
		repo.StargazersCount,
		repo.ReadmeContent,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("could not execute statement for repo %s: %w", repo.FullName, err)
	}

	return nil
}

// GetReposForSummarization retrieves repositories that have a README but no summary.
func GetReposForSummarization(db *sql.DB, limit int) ([]Repository, error) {
	query := `
		SELECT id, full_name, readme_content
		FROM repositories
		WHERE readme_content IS NOT NULL AND readme_content != ''
		AND (summary IS NULL OR summary = '')
		LIMIT ?;
	`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query repos for summarization: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		if err := rows.Scan(&repo.ID, &repo.FullName, &repo.ReadmeContent); err != nil {
			return nil, fmt.Errorf("could not scan repo row: %w", err)
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// UpdateRepoSummary updates the summary for a given repository.
func UpdateRepoSummary(db *sql.DB, repoID int64, summary string) error {
	query := `UPDATE repositories SET summary = ? WHERE id = ?;`
	_, err := db.Exec(query, summary, repoID)
	if err != nil {
		return fmt.Errorf("could not update summary for repo %d: %w", repoID, err)
	}
	return nil
}
