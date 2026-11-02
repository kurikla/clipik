package integration

import (
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/model"
)

func TestClipboard_E2E(t *testing.T) {
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

	c.SendClipboard(model.Clipboard{Data: []byte("hello from test")})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		list, err := c.History(10)

		if err != nil {
			t.Fatalf("history: %v", err)
		}

		if len(list) > 0 {
			if string(list[0].Data) != "hello from test" {
				t.Fatalf("data mismatch: %q", list[0].Data)
			}

			if list[0].Mime != "text/plain; charset=utf-8" {
				t.Fatalf("mime mismtach: %q", list[0].Mime)
			}

			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}
