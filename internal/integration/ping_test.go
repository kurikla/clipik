package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/dto"
	"github.com/kurikla/clipik/pkg/logger"
)

func TestPing(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	d := &FakeDiscoverer{Addr: h.ServerAddr, AutoAccept: true}

	st, err := client.NewSqliteStore(h.StorePath)
	if err != nil {
		t.Fatalf("client store: %v", err)
	}

	c, err := client.New(&client.Config{
		IdentityPath: h.ClientID,
		Format: dto.FormatMsgpack,
		SizeLimit: 64 << 20,
		HandshakeTimeout: 5 * time.Second,
		LogDir: filepath.Join(h.Dir, "client-logs"),
		LogLevel: logger.LogDebug,
		LogPrefix: "client",
		PingInterval: 200 * time.Millisecond,
		Store: st,
		Discoverer: d,
	})

	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer c.Stop()

	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitConnected(t, c, h.ServerAddr, 3*time.Second)

	time.Sleep(700 * time.Millisecond)

	conns := c.Connected()

	if len(conns) != 1 || conns[0] != h.ServerAddr {
		t.Fatalf("expected still connected to %s, got %v", h.ServerAddr, conns)
	}
}

func TestPing_ServerDownDisconnects(t *testing.T) {
	h := NewHarness(t)
	defer h.Cleanup()

	d := &FakeDiscoverer{Addr: h.ServerAddr, AutoAccept: true}

	st, err := client.NewSqliteStore(h.StorePath)
	if err != nil {
		t.Fatalf("client store: %v", err)
	}

	c, err := client.New(&client.Config{
		IdentityPath: h.ClientID,
		Format: dto.FormatMsgpack,
		SizeLimit: 64 << 20,
		HandshakeTimeout: 5 * time.Second,
		LogDir: filepath.Join(h.Dir, "client-logs"),
		LogLevel: logger.LogDebug,
		LogPrefix: "client",
		PingInterval: 200 * time.Millisecond,
		Store: st,
		Discoverer: d,
	})

	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer c.Stop()

	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitConnected(t, c, h.ServerAddr, 3*time.Second)

	h.Server.StopWithTimeout(3 * time.Second)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(c.Connected()) == 0 {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("client still connected after server down: %v", c.Connected())
}
