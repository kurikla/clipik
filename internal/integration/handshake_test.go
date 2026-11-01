package integration

import (
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/dto"
)

func TestHandshake_OK(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	d := &FakeDiscoverer{Addr: h.ServerAddr, AutoAccept: true}
	c := h.NewClient(d)
	defer c.Stop()

	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitConnected(t, c, h.ServerAddr, 3*time.Second)

	conns := c.Connected()
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %v", conns)
	}
}

func TestHandshake_BadFormat(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	_ = dto.Version
}
