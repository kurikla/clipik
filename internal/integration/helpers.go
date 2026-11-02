package integration

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/client"
	"github.com/kurikla/clipik/pkg/model"
)

func waitConnected(t *testing.T, c *client.Client, addr string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, a := range c.Connected() {
			if a == addr {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("client not connected to %s in %v", addr, timeout)
}

func waitDiscovered(t *testing.T, c *client.Client, addr string, timeout time.Duration) *model.Service {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, s := range c.Discovered() {
			if fmt.Sprintf("%s:%d", s.Ip, s.Port) == addr {
				return &s
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("client not discovered to %s in %v", addr, timeout)
	return nil
}

func portOf(addr string) int {
	i := strings.LastIndex(addr, ":")

	if i < 0 {
		return 0
	}

	p, _ := strconv.Atoi(addr[i+1:])
	return p
}
