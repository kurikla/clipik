package integration

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kurikla/clipik/pkg/client"
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

func portOf(addr string) int {
	i := strings.LastIndex(addr, ":")

	if i < 0 {
		return 0
	}

	p, _ := strconv.Atoi(addr[i+1:])
	return p
}
