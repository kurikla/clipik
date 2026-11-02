package integration

import (
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/protocol"
)

func TestHandshake_OK(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	d := &FakeDiscoverer{Addr: h.ServerAddr}
	c := h.NewClient(d)
	defer c.Stop()

	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	s := waitDiscovered(t, c, h.ServerAddr, 3*time.Second)
	c.Accept(*s)

	waitConnected(t, c, h.ServerAddr, 3*time.Second)

	conns := c.Connected()
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %v", conns)
	}
}

func TestHandshake_BadFormat(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	_ = protocol.Version
}
