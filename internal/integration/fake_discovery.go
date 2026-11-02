package integration

import (
	"log/slog"

	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/model"
)

type FakeDiscoverer struct {
	Addr string
}

func (f *FakeDiscoverer) SetLogger(*slog.Logger) error { return nil }

func (f *FakeDiscoverer) Start(observer *client.Observer) error {
	svc := model.Service{Ip: "127.0.0.1", Port: portOf(f.Addr), Name: "test"}
	observer.OnFound(svc)

	return nil
}

func (f *FakeDiscoverer) Stop() error { return nil }
