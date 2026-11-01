package client

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/kurikla/clipik-s/pkg/model"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(path string) (*SqliteStore, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("database %q does not exists", path)
	}

	db, err := sql.Open(
		"sqlite",
		path+"?mode=ro&_pragma=busy_timeout(5000)&_pragma=query_only(1)",
	)

	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxIdleConns(0)
	return &SqliteStore{db: db}, nil
}

func (s *SqliteStore) List(limit int) ([]model.Clipboard, error) {
	rows, err := s.db.Query(`
			SELECT id, mime, data, created_at
			FROM history
			ORDER BY id DESC
			LIMIT ?
	`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []model.Clipboard
	for rows.Next(){
		var c model.Clipboard
		if err := rows.Scan(&c.ID, &c.Mime, &c.Data, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	return out, rows.Err()
}

func (s *SqliteStore) Latest() (*model.Clipboard, error) {
	var c model.Clipboard
	err := s.db.QueryRow(`
			SELECT id, mime, data, created_at
			FROM history
			ORDER BY id DESC
			LIMIT 1
	`).Scan(&c.ID, &c.Mime, &c.Data, &c.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}
