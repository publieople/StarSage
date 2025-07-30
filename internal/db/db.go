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
	ETag            string
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
	// Main table for repository data
	const createRepoTableSQL = `
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER NOT NULL PRIMARY KEY,
		full_name TEXT NOT NULL UNIQUE,
		description TEXT,
		url TEXT,
		language TEXT,
		stargazers_count INTEGER,
		readme_content TEXT,
		summary TEXT,
		etag TEXT,
		last_synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// FTS5 virtual table for full-text search
	const createFtsTableSQL = `
	CREATE VIRTUAL TABLE IF NOT EXISTS repos_fts USING fts5(
		full_name,
		description,
		readme_content,
		content='repositories',
		content_rowid='id'
	);`

	// Triggers to keep the FTS table in sync with the repositories table
	const createTriggersSQL = `
	CREATE TRIGGER IF NOT EXISTS repos_ai AFTER INSERT ON repositories BEGIN
		INSERT INTO repos_fts(rowid, full_name, description, readme_content)
		VALUES (new.id, new.full_name, new.description, new.readme_content);
	END;
	CREATE TRIGGER IF NOT EXISTS repos_ad AFTER DELETE ON repositories BEGIN
		INSERT INTO repos_fts(repos_fts, rowid, full_name, description, readme_content)
		VALUES ('delete', old.id, old.full_name, old.description, old.readme_content);
	END;
	CREATE TRIGGER IF NOT EXISTS repos_au AFTER UPDATE ON repositories BEGIN
		INSERT INTO repos_fts(repos_fts, rowid, full_name, description, readme_content)
		VALUES ('delete', old.id, old.full_name, old.description, old.readme_content);
		INSERT INTO repos_fts(rowid, full_name, description, readme_content)
		VALUES (new.id, new.full_name, new.description, new.readme_content);
	END;
	`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(createRepoTableSQL); err != nil {
		return err
	}
	if _, err := tx.Exec(createFtsTableSQL); err != nil {
		return err
	}
	if _, err := tx.Exec(createTriggersSQL); err != nil {
		return err
	}

	return tx.Commit()
}

// UpsertRepository inserts or updates a single repository in the database.
func UpsertRepository(db *sql.DB, repo Repository) error {
	stmt, err := db.Prepare(`
		INSERT INTO repositories (id, full_name, description, url, language, stargazers_count, readme_content, etag, last_synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			full_name=excluded.full_name,
			description=excluded.description,
			url=excluded.url,
			language=excluded.language,
			stargazers_count=excluded.stargazers_count,
			readme_content=excluded.readme_content,
			etag=excluded.etag,
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
		repo.ETag,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("could not execute statement for repo %s: %w", repo.FullName, err)
	}

	return nil
}

// GetAllReposWithETags retrieves all repositories with their ID, ETag, and ReadmeContent.
func GetAllReposWithETags(db *sql.DB) ([]Repository, error) {
	query := `SELECT id, etag, readme_content FROM repositories;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not query repos for etags: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		var etag sql.NullString
		var readme sql.NullString
		if err := rows.Scan(&repo.ID, &etag, &readme); err != nil {
			return nil, fmt.Errorf("could not scan repo etag row: %w", err)
		}
		if etag.Valid {
			repo.ETag = etag.String
		}
		if readme.Valid {
			repo.ReadmeContent = readme.String
		}
		repos = append(repos, repo)
	}

	return repos, nil
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

// GetAllRepositories retrieves all repositories from the database.
func GetAllRepositories(db *sql.DB) ([]Repository, error) {
	query := `
		SELECT id, full_name, description, url, language, stargazers_count, summary, etag
		FROM repositories
		ORDER BY stargazers_count DESC;
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not query all repos: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		var desc, summary, etag sql.NullString
		if err := rows.Scan(
			&repo.ID,
			&repo.FullName,
			&desc,
			&repo.URL,
			&repo.Language,
			&repo.StargazersCount,
			&summary,
			&etag,
		); err != nil {
			return nil, fmt.Errorf("could not scan repo row: %w", err)
		}
		repo.Description = desc.String
		repo.Summary = summary.String
		repo.ETag = etag.String
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

// SearchRepositories performs a full-text search on the repositories.
func SearchRepositories(db *sql.DB, query string, limit int) ([]Repository, error) {
	// The snippet function highlights the search terms in the results.
	// The bm25 function provides relevancy ranking.
	// Use COALESCE to handle NULL summary values gracefully.
	searchSQL := `
		SELECT
			r.id,
			r.full_name,
			COALESCE(r.description, ''),
			r.url,
			r.language,
			r.stargazers_count,
			COALESCE(r.summary, ''),
			snippet(repos_fts, 1, '<b>', '</b>', '...', 15) as desc_snippet,
			bm25(repos_fts) as rank
		FROM repositories r
		JOIN repos_fts ON r.id = repos_fts.rowid
		WHERE repos_fts MATCH ?
		ORDER BY rank
	`
	var args []interface{}
	args = append(args, query)

	if limit > 0 {
		searchSQL += " LIMIT ?;"
		args = append(args, limit)
	} else {
		searchSQL += ";"
	}

	rows, err := db.Query(searchSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("could not execute search query: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		var descSnippet string
		var rank float64
		if err := rows.Scan(
			&repo.ID,
			&repo.FullName,
			&repo.Description,
			&repo.URL,
			&repo.Language,
			&repo.StargazersCount,
			&repo.Summary,
			&descSnippet,
			&rank,
		); err != nil {
			return nil, fmt.Errorf("could not scan search result row: %w", err)
		}
		// If the original description was empty, we can use the snippet as a fallback.
		if repo.Description == "" {
			repo.Description = descSnippet
		}
		repos = append(repos, repo)
	}

	return repos, nil
}
