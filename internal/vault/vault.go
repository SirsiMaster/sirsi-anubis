// Package vault implements a SQLite FTS5 context sandbox that stores large tool output
// outside the AI context window, queryable via full-text search with BM25 ranking.
//
// Subsumes the external Context Mode tool as native Go inside Pantheon.
package vault

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	// Pure-Go SQLite driver — no CGO required (Rule A3 compliance).
	_ "modernc.org/sqlite"
)

// Store manages the SQLite FTS5 context vault.
type Store struct {
	db   *sql.DB
	path string
}

// Entry is a single piece of sandboxed output.
type Entry struct {
	ID        int64  `json:"id"`
	Source    string `json:"source"`
	Tag       string `json:"tag"`
	Content   string `json:"content"`
	Tokens    int    `json:"tokens"`
	CreatedAt string `json:"createdAt"`
	Snippet   string `json:"snippet,omitempty"`
}

// SearchResult holds FTS5 query results.
type SearchResult struct {
	Query     string  `json:"query"`
	TotalHits int     `json:"totalHits"`
	Entries   []Entry `json:"entries"`
}

// StoreStats holds vault statistics.
type StoreStats struct {
	TotalEntries int            `json:"totalEntries"`
	TotalBytes   int64          `json:"totalBytes"`
	TotalTokens  int64          `json:"totalTokens"`
	OldestEntry  string         `json:"oldestEntry"`
	NewestEntry  string         `json:"newestEntry"`
	TagCounts    map[string]int `json:"tagCounts"`
}

// DefaultPath returns the default vault database path.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sirsi", "vault", "context.db")
}

// Open opens or creates the vault database at the given path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create vault dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open vault db: %w", err)
	}

	// Enable WAL mode for concurrent reads.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init vault schema: %w", err)
	}

	return &Store{db: db, path: path}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
		CREATE VIRTUAL TABLE IF NOT EXISTS vault_fts USING fts5(
			source,
			tag,
			content,
			tokenize = 'porter unicode61'
		);
		CREATE TABLE IF NOT EXISTS vault_meta (
			rowid INTEGER PRIMARY KEY,
			tokens INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now'))
		);
	`
	_, err := db.Exec(schema)
	return err
}

// Store sandboxes a piece of output into the vault.
func (s *Store) Store(source, tag, content string, tokens int) (*Entry, error) {
	res, err := s.db.Exec(
		"INSERT INTO vault_fts(source, tag, content) VALUES (?, ?, ?)",
		source, tag, content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert vault entry: %w", err)
	}
	rowid, _ := res.LastInsertId()

	_, err = s.db.Exec(
		"INSERT INTO vault_meta(rowid, tokens) VALUES (?, ?)",
		rowid, tokens,
	)
	if err != nil {
		return nil, fmt.Errorf("insert vault meta: %w", err)
	}

	return &Entry{
		ID:     rowid,
		Source: source,
		Tag:    tag,
		Tokens: tokens,
	}, nil
}

// Search performs an FTS5 full-text search with BM25 ranking.
func (s *Store) Search(query string, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
		SELECT f.rowid, f.source, f.tag,
			snippet(vault_fts, 2, '»', '«', '…', 40) as snip,
			m.tokens, m.created_at
		FROM vault_fts f
		JOIN vault_meta m ON m.rowid = f.rowid
		WHERE vault_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("vault search: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Source, &e.Tag, &e.Snippet, &e.Tokens, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan vault row: %w", err)
		}
		entries = append(entries, e)
	}

	return &SearchResult{
		Query:     query,
		TotalHits: len(entries),
		Entries:   entries,
	}, nil
}

// Get retrieves a specific entry by ID with full content.
func (s *Store) Get(id int64) (*Entry, error) {
	var e Entry
	err := s.db.QueryRow(`
		SELECT f.rowid, f.source, f.tag, f.content, m.tokens, m.created_at
		FROM vault_fts f
		JOIN vault_meta m ON m.rowid = f.rowid
		WHERE f.rowid = ?
	`, id).Scan(&e.ID, &e.Source, &e.Tag, &e.Content, &e.Tokens, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get vault entry %d: %w", id, err)
	}
	return &e, nil
}

// Stats returns vault statistics.
func (s *Store) Stats() (*StoreStats, error) {
	var stats StoreStats
	stats.TagCounts = make(map[string]int)

	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(LENGTH(content)), 0), COALESCE(SUM(m.tokens), 0)
		FROM vault_fts f
		JOIN vault_meta m ON m.rowid = f.rowid
	`).Scan(&stats.TotalEntries, &stats.TotalBytes, &stats.TotalTokens)
	if err != nil {
		return nil, fmt.Errorf("vault stats: %w", err)
	}

	_ = s.db.QueryRow("SELECT MIN(created_at) FROM vault_meta").Scan(&stats.OldestEntry)
	_ = s.db.QueryRow("SELECT MAX(created_at) FROM vault_meta").Scan(&stats.NewestEntry)

	rows, err := s.db.Query("SELECT tag, COUNT(*) FROM vault_fts GROUP BY tag")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag string
			var count int
			if rows.Scan(&tag, &count) == nil {
				stats.TagCounts[tag] = count
			}
		}
	}

	return &stats, nil
}

// Prune removes entries older than the given duration. Returns count of removed entries.
func (s *Store) Prune(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).UTC().Format("2006-01-02 15:04:05")

	// Get rowids to delete.
	rows, err := s.db.Query("SELECT rowid FROM vault_meta WHERE created_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("query prune targets: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return 0, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin prune tx: %w", err)
	}

	for _, id := range ids {
		tx.Exec("DELETE FROM vault_fts WHERE rowid = ?", id)
		tx.Exec("DELETE FROM vault_meta WHERE rowid = ?", id)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit prune: %w", err)
	}

	return len(ids), nil
}

// Close closes the vault database.
func (s *Store) Close() error {
	return s.db.Close()
}
