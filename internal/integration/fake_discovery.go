package integration

import (
	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/model"
)

type FakeDiscoverer struct {
	Addr string
}

func (f *FakeDiscoverer) Start(observer *client.Observer) error {
	svc := model.Service{Ip: "127.0.0.1", Port: portOf(f.Addr), Name: "test"}
	observer.OnFound(svc)

	return nil
}

func (f *FakeDiscoverer) Stop() error { return nil }
