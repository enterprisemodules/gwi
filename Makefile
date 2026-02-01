# gwi Makefile

BINARY_NAME=gwi
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build install clean test fmt lint

all: build

build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"
	@echo "Add to your shell config: eval \"\$$(gwi init zsh)\""

uninstall:
	rm -f $(HOME)/.local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...

# Cross-compilation targets
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-windows-amd64.exe .
