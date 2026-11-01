package integration

import (
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/model"
)

func TestDisconnect_OnLost(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	d := &FakeDiscoverer{Addr: h.ServerAddr, AutoAccept: true}
	c := h.NewClient(d)
	defer c.Stop()

	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitConnected(t, c, h.ServerAddr, 3*time.Second)

	_ = client.Observer{}

	_ = d.Stop()
	_ = model.Clipboard{}
	_ = model.Service{}
}
