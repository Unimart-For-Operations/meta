.PHONY: help init bootstrap status update drift tag hooks ci check test test-shell test-clean build install dev fmt vet completion

# Default target
.DEFAULT_GOAL := help

# Color output
CYAN   := \033[0;36m
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
RESET  := \033[0m
BOLD   := \033[1m

# Status indicators
PASS := \033[0;32m[pass]\033[0m
FAIL := \033[0;31m[fail]\033[0m
WARN := \033[0;33m[warn]\033[0m
INFO := \033[0;36m[info]\033[0m

# Use bash for pipefail support
SHELL := /usr/bin/env bash

# Submodules tracked by this org repo
SUBMODULES := $(shell git config --file .gitmodules --get-regexp path 2>/dev/null | awk '{print $$2}')

# Build variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/Unimart-For-Operations/meta/cmd.Version=$(VERSION) \
           -X github.com/Unimart-For-Operations/meta/cmd.GitCommit=$(COMMIT) \
           -X github.com/Unimart-For-Operations/meta/cmd.BuildDate=$(DATE)

# Delegate to unimart when available (installed by make init / make install)
HAS_UNIMART := $(shell command -v unimart 2>/dev/null)

help: ## Show this help message
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)Unimart-For-Operations/meta — Org Coordination$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@printf "$(BOLD)Setup:$(RESET)\n"
	@printf "  $(CYAN)init$(RESET)            Full setup: submodules → prerequisites → register → apply\n"
	@printf "  $(CYAN)bootstrap$(RESET)       Initialize submodules and install hooks\n"
	@printf "  $(CYAN)hooks$(RESET)           Install git hooks in all repos\n"
	@echo ""
	@printf "$(BOLD)CLI (unimart):$(RESET)\n"
	@printf "  $(CYAN)build$(RESET)           Build the unimart binary\n"
	@printf "  $(CYAN)install$(RESET)         Build and install to ~/.local/bin/unimart\n"
	@printf "  $(CYAN)dev$(RESET)             Run unimart from source (pass ARGS=\"...\")\n"
	@printf "  $(CYAN)fmt$(RESET)             Format Go source\n"
	@printf "  $(CYAN)vet$(RESET)             Run go vet\n"
	@printf "  $(CYAN)completion$(RESET)      Generate shell completion (COMPLETION_SHELL=zsh|bash|fish)\n"
	@echo ""
	@printf "$(BOLD)Submodule Management:$(RESET)\n"
	@printf "  $(CYAN)status$(RESET)          Show submodule state (dirty/ahead/behind)\n"
	@printf "  $(CYAN)update$(RESET)          Pull latest main for all submodules\n"
	@printf "  $(CYAN)drift$(RESET)           Check submodule pointers vs remote HEAD\n"
	@echo ""
	@printf "$(BOLD)Cross-Repo:$(RESET)\n"
	@printf "  $(CYAN)ci$(RESET)              Validate cross-repo contracts\n"
	@echo ""
	@printf "$(BOLD)Testing:$(RESET)\n"
	@printf "  $(CYAN)test$(RESET)            Run make init integration test in Linux container\n"
	@printf "  $(CYAN)test-shell$(RESET)      Drop into interactive shell in test container\n"
	@printf "  $(CYAN)test-clean$(RESET)      Tear down test containers and volumes\n"
	@echo ""
	@printf "$(BOLD)Release:$(RESET)\n"
	@printf "  $(CYAN)tag$(RESET)             Create org snapshot tag: make tag VERSION=v0.x.0\n"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ── CLI (unimart) ──────────────────────────────────────────────────────────

build: ## Build the unimart binary
	@go fmt ./...
	@go vet ./...
	@go build -ldflags "$(LDFLAGS)" -o unimart .
	@printf "  $(PASS) unimart built ($(VERSION))\n"

install: build ## Build and symlink to ~/.local/bin/unimart
	@mkdir -p ~/.local/bin
	@ln -sf "$(CURDIR)/unimart" ~/.local/bin/unimart
	@printf "  $(PASS) unimart installed → ~/.local/bin/unimart\n"

dev: ## Run unimart from source (pass ARGS="...")
	@go run -ldflags "$(LDFLAGS)" . $(ARGS)

fmt: ## Format Go source
	@go fmt ./...
	@printf "  $(PASS) formatted\n"

vet: ## Run go vet
	@go vet ./...
	@printf "  $(PASS) vet passed\n"

completion: ## Generate shell completion (COMPLETION_SHELL=zsh|bash|fish)
	@COMPLETION_SHELL="$${COMPLETION_SHELL:-$$(basename "$${SHELL:-zsh}")}"; \
	case "$$COMPLETION_SHELL" in \
		zsh) \
			DIR="$${COMPLETION_DIR:-$$HOME/.local/share/zsh/site-functions}"; \
			mkdir -p "$$DIR"; \
			go run . completion zsh > "$$DIR/_unimart"; \
			printf "  $(PASS) zsh completion → $$DIR/_unimart\n"; \
			;; \
		bash) \
			DIR="$${COMPLETION_DIR:-$$HOME/.local/share/bash-completion/completions}"; \
			mkdir -p "$$DIR"; \
			go run . completion bash > "$$DIR/unimart"; \
			printf "  $(PASS) bash completion → $$DIR/unimart\n"; \
			;; \
		fish) \
			DIR="$${COMPLETION_DIR:-$$HOME/.config/fish/completions}"; \
			mkdir -p "$$DIR"; \
			go run . completion fish > "$$DIR/unimart.fish"; \
			printf "  $(PASS) fish completion → $$DIR/unimart.fish\n"; \
			;; \
		*) \
			printf "  $(FAIL) unsupported shell: $$COMPLETION_SHELL (use zsh, bash, or fish)\n"; \
			exit 1; \
	esac

# ── Setup ──────────────────────────────────────────────────────────────────

init: ## Compatibility wrapper for the CLI onboarding flow
	@if command -v unimart >/dev/null 2>&1; then \
		unimart deli bootstrap; \
	else \
		go run . deli bootstrap; \
	fi

bootstrap: ## Compatibility wrapper for the CLI onboarding flow
	@if command -v unimart >/dev/null 2>&1; then \
		unimart deli bootstrap; \
	else \
		go run . deli bootstrap; \
	fi

hooks: ## Show hook status (hooks are Nix-managed — see ADR-005)
	@printf "$(BOLD)Git Hook Gate System$(RESET)\n"
	@printf "  $(INFO) Hooks are globally managed via Nix (cmdr git module)\n"
	@printf "  $(INFO) Deploy: unimart deli switch\n"
	@printf "  $(INFO) Extend: add .githooks/<hook-name> to any repo\n"
	@printf "  $(INFO) Docs:   docs/Architecture/adr/005-git-hook-gates.md\n"
	@echo ""
	@printf "  $(BOLD)Gates:$(RESET)\n"
	@printf "    pre-commit:  nix fmt, go fmt, go vet, gitleaks\n"
	@printf "    commit-msg:  conventional commit, DCO, Changes, Executive Summary\n"
	@printf "    pre-push:    go build, go test, nix flake check\n"

# ── Submodule Management ──────────────────────────────────────────────────

status: ## Compatibility wrapper for stockroom status
	@if command -v unimart >/dev/null 2>&1; then \
		unimart stockroom status; \
	else \
		printf "$(BOLD)stockroom status is available via: unimart stockroom check$(RESET)\n"; \
		printf "$(INFO) build unimart first with: make build$(RESET)\n"; \
	fi

update: ## Compatibility wrapper for stockroom update
	@if command -v unimart >/dev/null 2>&1; then \
		unimart stockroom update; \
	else \
		printf "$(BOLD)stockroom update is available via the CLI$(RESET)\n"; \
		printf "$(INFO) build unimart first with: make build$(RESET)\n"; \
	fi

drift: ## Compatibility wrapper for stockroom drift
	@if command -v unimart >/dev/null 2>&1; then \
		unimart stockroom drift; \
	else \
		printf "$(BOLD)stockroom drift is available via the CLI$(RESET)\n"; \
		printf "$(INFO) build unimart first with: make build$(RESET)\n"; \
	fi


# ── Cross-Repo ─────────────────────────────────────────────────────────────

ci: ## Compatibility wrapper for stockroom check
	@if command -v unimart >/dev/null 2>&1; then \
		unimart stockroom check; \
	else \
		printf "$(BOLD)stockroom check is available via the CLI$(RESET)\n"; \
		printf "$(INFO) build unimart first with: make build$(RESET)\n"; \
	fi

check: ci ## Alias for ci

# ── Testing (container-based) ──────────────────────────────────────────────

# Compose file location
COMPOSE_FILE := containers/podman-compose.yml
COMPOSE_CMD  := podman-compose -f $(COMPOSE_FILE)

test: ## Run the CLI-first smoke test in a Linux container
	@printf "$(BOLD)Building test container...$(RESET)\n"
	@$(COMPOSE_CMD) up -d --build
	@printf "$(BOLD)Running CLI smoke test (this takes several minutes)...$(RESET)\n"
	@$(COMPOSE_CMD) exec -T init-test bash /workspace/containers/test-init.sh; \
	EXIT=$$?; \
	printf "\n$(BOLD)Tearing down container...$(RESET)\n"; \
	$(COMPOSE_CMD) down -v >/dev/null 2>&1; \
	exit $$EXIT

test-shell: ## Drop into interactive shell in test container
	@printf "$(BOLD)Building test container...$(RESET)\n"
	@$(COMPOSE_CMD) up -d --build
	@printf "$(BOLD)Entering container shell...$(RESET)\n"
	@$(COMPOSE_CMD) exec init-test /bin/bash

test-clean: ## Tear down test containers and volumes
	@printf "Cleaning up test containers and volumes...\n"
	@$(COMPOSE_CMD) down -v 2>/dev/null || true
	@printf "$(GREEN)[pass] Cleanup complete$(RESET)\n"

# ── Release ────────────────────────────────────────────────────────────────

tag: ## Create org snapshot tag: make tag VERSION=v0.x.0
	@if [ -z "$(VERSION)" ]; then \
		printf "$(YELLOW)Usage: make tag VERSION=v0.x.0$(RESET)\n"; \
		exit 1; \
	fi
	@if git tag -l "$(VERSION)" | grep -q .; then \
		printf "$(RED)Tag $(VERSION) already exists$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BOLD)Creating org snapshot tag: $(VERSION)$(RESET)\n"
	@echo ""
	@# Build tag message with submodule versions
	@MSG="Org snapshot $(VERSION)\n\nSubmodule state:\n"; \
	for mod in $(SUBMODULES); do \
		SHORT=$$(git -C "$$mod" rev-parse --short HEAD 2>/dev/null); \
		BRANCH=$$(git -C "$$mod" rev-parse --abbrev-ref HEAD 2>/dev/null); \
		TAG=$$(git -C "$$mod" describe --tags --exact-match HEAD 2>/dev/null || echo "-"); \
		MSG="$$MSG  $$mod  $$BRANCH@$$SHORT"; \
		if [ "$$TAG" != "-" ]; then MSG="$$MSG ($$TAG)"; fi; \
		MSG="$$MSG\n"; \
	done; \
	printf "$$MSG" | git tag -a "$(VERSION)" -F -; \
	printf "$(GREEN)Tagged $(VERSION)$(RESET)\n"; \
	echo ""; \
	git tag -l "$(VERSION)" -n99; \
	echo ""; \
	printf "Push with: git push origin $(VERSION)\n"
