package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"star-sage/internal/gh"
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

// SaveRepositories saves a slice of repositories to the database.
// It uses a transaction for efficiency and data integrity.
func SaveRepositories(db *sql.DB, repos []gh.GHRepo) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	stmt, err := tx.Prepare(`
		INSERT INTO repositories (id, full_name, description, url, language, stargazers_count, last_synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			full_name=excluded.full_name,
			description=excluded.description,
			url=excluded.url,
			language=excluded.language,
			stargazers_count=excluded.stargazers_count,
			last_synced_at=excluded.last_synced_at;
	`)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, repo := range repos {
		_, err := stmt.Exec(
			repo.ID,
			repo.FullName,
			repo.Description,
			repo.HTMLURL,
			repo.Language,
			repo.StargazersCount,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("could not execute statement for repo %s: %w", repo.FullName, err)
		}
	}

	return tx.Commit()
}
