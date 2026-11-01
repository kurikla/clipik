package server

import "github.com/kurikla/clipik/pkg/model"

type Store interface {
	Save(c model.Clipboard) (*model.Clipboard, error)
	Trim(limit int) (int64, error)
	Close() error
}
