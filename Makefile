BINARY   := bin/server
PKG      := ./...
COVER    := coverage.out
# Thin I/O adapters with no unit tests: process wiring and the Firestore client.
COVER_EXCLUDE := /cmd/server/|/store/firestore.go
MIN_COV  ?= 80
IMAGE    ?= posse-utils
GOFLAGS  ?=

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Sync go.mod and go.sum
	go mod tidy

.PHONY: fmt
fmt: ## Format code
	golangci-lint fmt

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: vet
vet: ## Run go vet
	go vet $(PKG)

.PHONY: test
test: ## Run unit tests with race detector and coverage
	go test -race -shuffle=on -covermode=atomic -coverprofile=$(COVER) $(PKG)
	grep -Ev '$(COVER_EXCLUDE)' $(COVER) > $(COVER).tmp && mv $(COVER).tmp $(COVER)

.PHONY: cover
cover: test ## Print coverage per function
	go tool cover -func=$(COVER)

.PHONY: cover-html
cover-html: test ## Open HTML coverage report
	go tool cover -html=$(COVER)

.PHONY: cover-check
cover-check: test ## Fail if total coverage is below MIN_COV (default 80)
	@total=$$(go tool cover -func=$(COVER) | awk '/^total:/ {sub("%","",$$3); print $$3}'); \
	echo "total coverage: $$total% (min $(MIN_COV)%)"; \
	awk -v t="$$total" -v m="$(MIN_COV)" 'BEGIN {exit !(t+0 >= m+0)}'

.PHONY: build
build: ## Build the server binary into bin/
	CGO_ENABLED=0 go build $(GOFLAGS) -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/server

.PHONY: run
run: ## Run locally (in-memory store, mail logged). Override env as needed
	BASE_URL=$${BASE_URL:-http://localhost:8080} \
	SITE_URL=$${SITE_URL:-http://localhost:4321} \
	TOKEN_SECRET=$${TOKEN_SECRET:-dev-only-secret-dev-only-secret-0000} \
	go run ./cmd/server

.PHONY: check
check: vet lint cover-check ## Everything CI should run

.PHONY: docker-build
docker-build: ## Build the container image
	docker build -t $(IMAGE) .

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -rf bin $(COVER)
