GO_TEST_FLAGS ?= -v

build:
	go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o bin/clipik-s ./cmd/clipik-s

test-integration:
	go test ./internal/integration/... -count=1 $(GO_TEST_FLAGS)

test-all: test-integration
