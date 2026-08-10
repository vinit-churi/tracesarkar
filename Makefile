# TraceSarkar
#
# The repository is currently documentation and design. Targets marked
# "not implemented" will be filled in as code lands per docs/05-delivery/01-roadmap.md.

.DEFAULT_GOAL := help
SHELL := /usr/bin/env bash

# ---------------------------------------------------------------- docs (live)

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
	done < <(find . -name '*.md' -not -path './.git/*'); \
	if [ $$fail -eq 0 ]; then echo "links: ok"; fi; \
	exit $$fail

.PHONY: docs-hygiene
docs-hygiene: ## No tabs, no trailing whitespace in markdown
	@if grep -rIl "$$(printf '\t')" --include='*.md' . ; then echo "tabs found"; exit 1; fi
	@if grep -rIn ' $$' --include='*.md' . ; then echo "trailing whitespace found"; exit 1; fi
	@echo "hygiene: ok"

.PHONY: docs-language
docs-language: ## No accusatory language in product copy or templates
	@if grep -rInE '\b(scam|loot|shameful|caught red-handed)\b' docs/templates docs/02-product ; then \
	  echo "accusatory language found; see docs/00-overview/04-naming-and-brand.md"; exit 1; fi
	@echo "language: ok"

.PHONY: docs-unverified
docs-unverified: ## List every claim still marked as unverified
	@grep -rn '⚠️' docs/ | sed 's/^/  /' || true

# ------------------------------------------------------- application (pending)

.PHONY: dev
dev: ## Start local dependencies (postgres+postgis, redis, minio)
	@echo "not implemented yet — see docs/05-delivery/03-backlog.md (repo scaffold)"
	@exit 1

.PHONY: migrate
migrate: ## Apply database migrations
	@echo "not implemented yet"
	@exit 1

.PHONY: seed
seed: ## Load synthetic MMR development data (never real citizen data)
	@echo "not implemented yet"
	@exit 1

.PHONY: test
test: ## Run the Go test suite
	@echo "not implemented yet"
	@exit 1

.PHONY: lint
lint: ## Run golangci-lint, go vet, gosec
	@echo "not implemented yet"
	@exit 1

.PHONY: eval
eval: ## Run the AI evaluation suites
	@echo "not implemented yet — see docs/03-architecture/04-ai-pipeline.md"
	@exit 1

# ------------------------------------------------------------------------ help

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
