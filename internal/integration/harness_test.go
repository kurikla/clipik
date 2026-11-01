package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/dto"
	"github.com/kurikla/clipik/pkg/logger"
	"github.com/kurikla/clipik/pkg/server"
)

type Harness struct {
	t          *testing.T
	Server     *server.Server
	ServerAddr string
	ServerID   string
	ClientID   string
	StorePath  string
	Dir		   string
}

func NewHarness(t *testing.T) *Harness {
	t.Helper()

	dir := t.TempDir()

	h := &Harness{
		t:         t,
		ServerID:  filepath.Join(dir, "server.identity"),
		ClientID:  filepath.Join(dir, "client.identity"),
		StorePath: filepath.Join(dir, "clipik.db"),
		Dir:       dir,
	}

	st, err := server.NewSqliteStore(h.StorePath)

	if err != nil {
		t.Fatalf("store: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	h.ServerAddr = fmt.Sprintf("127.0.0.1:%d", port)

	srv, err := server.New(
		&server.Config{
			Port: port,
			Interfaces: nil,
			FilesDir: filepath.Join(dir, "files"),
			LogDir: filepath.Join(dir, "server-logs"),
			LogLevel: logger.LogDebug,
			LogPrefix: "server",
			LogToStdout: logToStdout(),
			SizeLimit: 64 << 20,
			HandshakeTimeout: 5 * time.Second,
			TrimInterval: 500 * time.Millisecond,
			TrimLimit: 100,
			IdentityPath: h.ServerID,
			Store: st,
			Discoverer: server.Noop{},
		},
	)

	if err != nil {
		t.Fatalf("server: %v", err)
	}

	h.Server = srv

	go func () {
		if err := srv.Start(); err != nil {
			t.Logf("server start: %v", err)
		}
	}()

	waitPort(t, h.ServerAddr, 3*time.Second)

	return h
}

func (h *Harness) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = h.Server.StopWithTimeout(3 * time.Second)
	_ = ctx
}

func (h *Harness) NewClient(d client.Discoverer) *client.Client {
	h.t.Helper()

	st, err := client.NewSqliteStore(h.StorePath)
	if err != nil {
		h.t.Fatalf("client store: %v", err)
	}

	c, err := client.New(&client.Config{
		IdentityPath: h.ClientID,
		Format: dto.FormatMsgpack,
		SizeLimit: 64 << 20,
		HandshakeTimeout: 5 * time.Second,
		LogDir: filepath.Join(h.Dir, "client-logs"),
		LogLevel: logger.LogError,
		LogPrefix: "client",
		LogToStdout: logToStdout(),
		Store: st,
		Discoverer: d,
	})

	if err != nil {
		h.t.Fatalf("client: %v", err)
	}

	return c
}

func waitPort(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)

		if err == nil {
			conn.Close()
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("port %s not ready in %v", addr, timeout)
}

func init() {
	_ = os.MkdirAll
}
