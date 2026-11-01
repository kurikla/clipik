package client

import (
	"github.com/kurikla/clipik/pkg/model"
	"github.com/google/uuid"
)

type Store interface {
	Get(id uuid.UUID) (*model.Clipboard, error)
	List(limit int) ([]model.Clipboard, error)
	Latest() (*model.Clipboard, error)
	Close() error
}
