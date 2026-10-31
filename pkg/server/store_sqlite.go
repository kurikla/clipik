package server


import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kurikla/clipik-s/pkg/model"
	_ "modernc.org/sqlite"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(path string) (*SqliteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", path+
		"?_pragma=journal_mode(WAL)"+
		"&_pragma=busy_timeout(5000)"+
		"&pragma=synchronous(NORMAL)")

	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxIdleConns(1)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			mime       TEXT NOT NULL,
			data       BLOB NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_history_created_at ON history (created_at DESC);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &SqliteStore{db: db}, nil
}

func (s *SqliteStore) Save(c model.Clipboard) (*model.Clipboard, error) {
	err := s.db.QueryRow(
		`INSERT INTO history (mime, data) VALUES (?, ?) RETURNING id, created_at`,
		c.Mime,
		c.Data,
	).Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *SqliteStore) Trim(limit int) (int64, error) {
	res, err := s.db.Exec(`DELETE FROM history WHERE id NOT IN (SELECT id FROM history ORDER BY id DESC LIMIT ?)`, limit)

	if err != nil {
		return 0, err
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}
