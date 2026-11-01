package integration

import (
	"flag"
	"os"
	"testing"
)

var logStdout = flag.Bool("logstdout", false, "enable logger stdout in integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func logToStdout() bool {
	return *logStdout
}
