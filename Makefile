.PHONY: all build build-dashboard build-go run clean release test test-verbose dev-backend dev-frontend lint

VERSION ?= dev
BINARY  := fracture
GOFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

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

## test: run Go tests
test:
	go test ./... -v -race -timeout 60s

## test-verbose: run Go tests with verbose output
test-verbose:
	go test -v ./...

## dev-backend: run Go server in dev mode (no dashboard build)
dev-backend:
	go run .

## dev-frontend: run React dashboard dev server
dev-frontend:
	cd dashboard && pnpm dev

## lint: run Go vet
lint:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
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
