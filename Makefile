BINARY  := speeder
MODULE  := github.com/mhdiiilham/speeder
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

TARGETS := \
	linux/amd64 linux/arm64 linux/arm \
	darwin/amd64 darwin/arm64 \
	windows/amd64 windows/arm64

.PHONY: build install test coverage lint fmt clean release

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

install:
	go install $(LDFLAGS) .

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html
	@go tool cover -func=cover.out | tail -1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin/ cover.out cover.html

release: clean
	$(foreach TARGET,$(TARGETS), \
		GOOS=$(word 1,$(subst /, ,$(TARGET))) \
		GOARCH=$(word 2,$(subst /, ,$(TARGET))) \
		go build $(LDFLAGS) \
			-o bin/$(BINARY)-$(subst /,-,$(TARGET))$(if $(findstring windows,$(TARGET)),.exe,) \
			. ;)
	@echo "Built: $(TARGETS)"
