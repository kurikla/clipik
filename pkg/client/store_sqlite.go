package client

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/kurikla/clipik/pkg/model"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(path string) (*SqliteStore, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("database %q does not exists", path)
	}

	shmExists := fileExists(path + "-shm")
	walExists := fileExists(path + "-wal")

	var dsn string
	if !shmExists || !walExists {
		dsn = fmt.Sprintf("file:%s?immutable=1", path)
	} else {
		dsn = fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)&_pragma=query_only(1)", path)
	}

	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxIdleConns(0)

	return &SqliteStore{db: db}, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (s *SqliteStore) Get(id uuid.UUID) (*model.Clipboard, error) {
	var c model.Clipboard

	err := s.db.QueryRow(`
			SELECT id, mime, data, created_at
			FROM history
			WHERE id = ?
	`, id).Scan(&c.ID, &c.Mime, &c.Data, &c.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *SqliteStore) List(limit int) ([]model.Clipboard, error) {
	rows, err := s.db.Query(`
			SELECT id, mime, data, created_at
			FROM history
			ORDER BY created_at DESC, rowid DESC
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
