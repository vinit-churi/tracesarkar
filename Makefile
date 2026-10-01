# TraceSarkar
#
# The repository is currently documentation and design. Targets marked
# "not implemented" will be filled in as code lands per docs/05-delivery/01-roadmap.md.

.DEFAULT_GOAL := help
SHELL := /usr/bin/env bash

# ---------------------------------------------------------------- docs (live)

# Only our own markdown. node_modules and build output arrived with the Flutter
# client and are full of other people's broken links.
DOCS_PRUNE = -path './.git' -o -path '*/node_modules' -o -path '*/build' \
             -o -path '*/.dart_tool' -o -path './app/ios' -o -path './app/android'
docs_files = find . \( $(DOCS_PRUNE) \) -prune -o -name '*.md' -print

.PHONY: docs-check
docs-check: docs-links docs-hygiene docs-language ## Run all documentation checks

.PHONY: docs-links
docs-links: ## Verify every relative markdown link resolves
	@fail=0; \
	while IFS= read -r file; do \
	  dir=$$(dirname "$$file"); \
	  while IFS= read -r target; do \
	    path="$${target%%\#*}"; \
	    [ -z "$$path" ] && continue; \
	    if [ ! -e "$$dir/$$path" ]; then echo "BROKEN: $$file -> $$target"; fail=1; fi; \
	  done < <(grep -oE '\]\([^)]+\)' "$$file" | sed -E 's/^\]\(//; s/\)$$//' | grep -vE '^(https?:|mailto:|\#)'); \
	done < <($(docs_files)); \
	if [ $$fail -eq 0 ]; then echo "links: ok"; fi; \
	exit $$fail

.PHONY: docs-hygiene
docs-hygiene: ## No tabs, no trailing whitespace in markdown
	@if $(docs_files) | xargs grep -Il "$$(printf '\t')" ; then echo "tabs found"; exit 1; fi
	@if $(docs_files) | xargs grep -In ' $$' ; then echo "trailing whitespace found"; exit 1; fi
	@echo "hygiene: ok"

.PHONY: docs-language
docs-language: ## No accusatory language in product copy or templates
	@if grep -rInE '\b(scam|loot|shameful|caught red-handed)\b' docs/templates docs/02-product ; then \
	  echo "accusatory language found; see docs/00-overview/04-naming-and-brand.md"; exit 1; fi
	@echo "language: ok"

.PHONY: docs-unverified
docs-unverified: ## List every claim still marked as unverified
	@grep -rn '⚠️' docs/ | sed 's/^/  /' || true

# ------------------------------------------------------------- application

.PHONY: build
build: ## Build the binaries into bin/
	@mkdir -p bin
	go build -o bin/ ./cmd/...
	@echo "built: $$(ls bin/)"

.PHONY: api
api: ## Serve the API and the field kit (needs API_TOKEN)
	go run ./cmd/api serve

.PHONY: migrate
migrate: ## Apply database migrations
	go run ./cmd/ingest migrate

.PHONY: snapshot
snapshot: ## Snapshot every schedulable source (Phase 0 P1)
	go run ./cmd/ingest run

.PHONY: watch
watch: ## Archive newly published Government Resolutions (Phase 0 P2)
	go run ./cmd/ingest watch

.PHONY: all-collect
all-collect: ## Snapshot and watch, exactly as the scheduled job does
	go run ./cmd/ingest all

.PHONY: status
status: ## Per-endpoint collection health
	go run ./cmd/ingest status

.PHONY: changes
changes: ## Changes observed in published works data
	go run ./cmd/ingest changes

.PHONY: test
test: ## Run the Go test suite (offline)
	go test ./... -count=1

.PHONY: test-live
test-live: ## Run the tests that touch the real bucket and database
	TRACESARKAR_LIVE=1 go test ./... -count=1

.PHONY: lint
lint: ## gofmt check and go vet
	@unformatted=$$(gofmt -l ./cmd ./internal); \
	if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi
	go vet ./...
	@echo "lint: ok"

.PHONY: seed
seed: ## Load synthetic MMR development data (never real citizen data)
	@echo "not implemented yet"
	@exit 1

.PHONY: eval
eval: ## Run the AI evaluation suites
	@echo "not implemented yet — see docs/03-architecture/04-ai-pipeline.md"
	@exit 1

# ------------------------------------------------------------------------- app

# The client is built with its server address and OAuth client id compiled in.
# Both come from .env so neither is ever retyped from memory — an API_BASE
# guessed wrong produces a client that silently talks to nothing.
FLUTTER ?= $(shell command -v flutter 2>/dev/null || echo $(HOME)/flutter-sdk/bin/flutter)

# Only these two keys are lifted out of .env. A whole `include .env` would have
# make try to expand the database password.
dotenv = $(shell sed -n 's/^$(1)=//p' .env 2>/dev/null | head -1)
API_BASE ?= $(call dotenv,API_BASE)
GOOGLE_CLIENT_ID ?= $(call dotenv,GOOGLE_CLIENT_ID)
APP_DEFINES = --dart-define=API_BASE=$(API_BASE) \
              --dart-define=GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID)

.PHONY: app-test
app-test: ## Run the Flutter tests
	cd app && $(FLUTTER) test

.PHONY: app-web
app-web: ## Build the Flutter web release
	@test -n "$(API_BASE)" || { echo "API_BASE unset — source .env"; exit 1; }
	cd app && $(FLUTTER) build web --release $(APP_DEFINES)

.PHONY: app-deploy
app-deploy: app-web ## Build and deploy the web client to Cloudflare Workers
	cd app && npx wrangler deploy

.PHONY: app-apk
app-apk: ## Build a release APK for sideloading
	@test -n "$(API_BASE)" || { echo "API_BASE unset — source .env"; exit 1; }
	cd app && $(FLUTTER) build apk --release $(APP_DEFINES)

# ------------------------------------------------------------------------ help

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
