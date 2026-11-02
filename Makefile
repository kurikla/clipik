GO_TEST_FLAGS ?= -v

build-s:
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty)" -o bin/clipik-s ./cmd/clipik-s

.PHONY: build-s

test-integration:
	go test ./internal/integration/... -count=1 $(GO_TEST_FLAGS)

.PHONY: test-integration

test-all: test-integration

.PHOY: test-allN
