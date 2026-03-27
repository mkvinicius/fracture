.PHONY: all build build-dashboard build-go run dev test lint fmt tidy coverage clean release release-snapshot install-goreleaser install-lint help

VERSION ?= dev
BINARY  := fracture
GOFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

## help: print this help message
help:
	@echo "FRACTURE v2.5.0 — available targets:"
	@sed -n 's/^## /  /p' $(MAKEFILE_LIST)

## all: build dashboard + Go binary
all: build

## build: build dashboard then embed into Go binary
build: build-dashboard build-go

## build-dashboard: compile React dashboard to dashboard/dist/
build-dashboard:
	@echo "→ Building dashboard..."
	cd dashboard && pnpm install --frozen-lockfile && pnpm build

## build-go: compile Go binary (embeds dashboard/dist)
build-go:
	@echo "→ Building Go binary..."
	go build $(GOFLAGS) -o $(BINARY) .

## run: build and run locally (opens browser at localhost:3000)
run: build
	@echo "→ Starting FRACTURE..."
	./$(BINARY)

## dev: run dashboard dev server + Go server in parallel
dev:
	@echo "→ Starting dev mode..."
	@trap 'kill %1 %2' INT; \
	  (cd dashboard && pnpm dev) & \
	  (go run . ) & \
	  wait

## test: run Go tests with race detector
test:
	go test ./... -v -race -timeout 60s

## lint: run golangci-lint (install with make install-lint)
lint:
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

## tidy: tidy and verify Go module dependencies
tidy:
	go mod tidy
	go mod verify

## coverage: run tests and open HTML coverage report
coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "→ Coverage report: coverage.html"

## clean: remove build artefacts and coverage files
clean:
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf dashboard/dist

## release: build release binaries for all platforms via GoReleaser
release:
	goreleaser release --clean

## release-snapshot: build without publishing (for testing)
release-snapshot:
	goreleaser release --snapshot --clean

## install-goreleaser: install GoReleaser
install-goreleaser:
	go install github.com/goreleaser/goreleaser/v2@latest

## install-lint: install golangci-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
