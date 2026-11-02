package client

import "log/slog"

type Discoverer interface {
	SetLogger(logger *slog.Logger) error
	Start(observer *Observer) error
	Stop() error
}
