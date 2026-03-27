.PHONY: all build build-dashboard build-go build-mac-app run clean release test setup

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

## build-mac-app: build FRACTURE.app + DMG for macOS (requires macOS)
build-mac-app:
	make build
	chmod +x scripts/build-mac-app.sh
	./scripts/build-mac-app.sh $(VERSION)

## setup: create .env from .env.example if it doesn't exist yet
setup:
	@cp -n .env.example .env && echo ".env criado — configure suas chaves" || echo ".env já existe"

## install-goreleaser: install GoReleaser
install-goreleaser:
	go install github.com/goreleaser/goreleaser/v2@latest
