package notify

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store manages the notification history SQLite database.
type Store struct {
	db *sql.DB
}

// Notification is a single history entry.
type Notification struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source"`
	Action     string    `json:"action"`
	Severity   string    `json:"severity"`
	Summary    string    `json:"summary"`
	Details    string    `json:"details,omitempty"`
	DurationMs int64     `json:"duration_ms"`
}

// DefaultPath returns the default notification store path.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sirsi", "notifications.db")
}

// Open opens or creates the notification store.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create notify dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open notify db: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if err := initNotifySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init notify schema: %w", err)
	}

	return &Store{db: db}, nil
}

func initNotifySchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS notifications (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp   TEXT    NOT NULL DEFAULT (datetime('now')),
			source      TEXT    NOT NULL,
			action      TEXT    NOT NULL,
			severity    TEXT    NOT NULL DEFAULT 'info',
			summary     TEXT    NOT NULL,
			details     TEXT,
			duration_ms INTEGER DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_notifications_timestamp ON notifications(timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_notifications_source ON notifications(source);
	`
	_, err := db.Exec(schema)
	return err
}

// Record stores a notification and fires a macOS toast.
// This is the primary entry point — combines storage + toast in one call.
func (s *Store) Record(n Notification) error {
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now()
	}

	_, err := s.db.Exec(
		`INSERT INTO notifications (timestamp, source, action, severity, summary, details, duration_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		n.Timestamp.Format(time.RFC3339), n.Source, n.Action,
		n.Severity, n.Summary, n.Details, n.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("record notification: %w", err)
	}

	// Fire toast — non-blocking, best-effort.
	icon := SeverityIcon(n.Severity)
	Toast(
		fmt.Sprintf("Sirsi %s %s", icon, n.Source),
		n.Summary,
	)

	return nil
}

// Recent returns the last N notifications, newest first.
func (s *Store) Recent(limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.query("SELECT id, timestamp, source, action, severity, summary, details, duration_ms FROM notifications ORDER BY timestamp DESC LIMIT ?", limit)
}

// BySource returns notifications filtered by source deity.
func (s *Store) BySource(source string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.query("SELECT id, timestamp, source, action, severity, summary, details, duration_ms FROM notifications WHERE source = ? ORDER BY timestamp DESC LIMIT ?", source, limit)
}

// BySeverity returns notifications filtered by severity.
func (s *Store) BySeverity(severity string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.query("SELECT id, timestamp, source, action, severity, summary, details, duration_ms FROM notifications WHERE severity = ? ORDER BY timestamp DESC LIMIT ?", severity, limit)
}

// Since returns notifications after the given time.
func (s *Store) Since(t time.Time, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.query("SELECT id, timestamp, source, action, severity, summary, details, duration_ms FROM notifications WHERE timestamp > ? ORDER BY timestamp DESC LIMIT ?", t.Format(time.RFC3339), limit)
}

func (s *Store) query(q string, args ...interface{}) ([]Notification, error) {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var results []Notification
	for rows.Next() {
		var n Notification
		var ts string
		var details sql.NullString
		if err := rows.Scan(&n.ID, &ts, &n.Source, &n.Action, &n.Severity, &n.Summary, &details, &n.DurationMs); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		n.Timestamp, _ = time.Parse(time.RFC3339, ts)
		if details.Valid {
			n.Details = details.String
		}
		results = append(results, n)
	}
	return results, nil
}

// Prune deletes notifications older than the given duration.
func (s *Store) Prune(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)
	result, err := s.db.Exec("DELETE FROM notifications WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune notifications: %w", err)
	}
	return result.RowsAffected()
}

// Clear deletes all notifications.
func (s *Store) Clear() (int64, error) {
	result, err := s.db.Exec("DELETE FROM notifications")
	if err != nil {
		return 0, fmt.Errorf("clear notifications: %w", err)
	}
	return result.RowsAffected()
}

// Count returns the total number of notifications.
func (s *Store) Count() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM notifications").Scan(&count)
	return count, err
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
