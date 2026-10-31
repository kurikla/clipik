build:
	go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o bin/clipik-s ./cmd/clipik-s
