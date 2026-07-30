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
LDFLAGS := -X github.com/idpbuilder/meta/cmd.Version=$(VERSION) \
           -X github.com/idpbuilder/meta/cmd.GitCommit=$(COMMIT) \
           -X github.com/idpbuilder/meta/cmd.BuildDate=$(DATE)

# Delegate to unimart when available (installed by make init / make install)
HAS_UNIMART := $(shell command -v unimart 2>/dev/null)

help: ## Show this help message
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)idpbuilder/meta — Org Coordination$(RESET)\n"
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

init: ## Full setup: submodules → prerequisites → register → apply → verify
	@bash scripts/setup.sh

bootstrap: ## Initialize submodules and install hooks across all repos
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)idpbuilder — Bootstrap$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@printf "$(BOLD)[1/3] Initializing submodules$(RESET)\n"
	@git submodule update --init --recursive
	@printf "  $(PASS) Submodules initialized\n"
	@echo ""
	@printf "$(BOLD)[2/3] Verifying submodule remotes$(RESET)\n"
	@ERRORS=0; \
	for mod in $(SUBMODULES); do \
		if [ -d "$$mod/.git" ] || [ -f "$$mod/.git" ]; then \
			REMOTE=$$(git -C "$$mod" remote get-url origin 2>/dev/null); \
			if echo "$$REMOTE" | grep -q 'github.com[:/]idpbuilder/'; then \
				printf "  $(PASS) $$mod → $$REMOTE\n"; \
			else \
				printf "  $(FAIL) $$mod → $$REMOTE (expected github.com:idpbuilder/)\n"; \
				ERRORS=$$((ERRORS + 1)); \
			fi; \
		else \
			printf "  $(FAIL) $$mod — not a git repo\n"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	done; \
	if [ $$ERRORS -gt 0 ]; then \
		printf "\n$(RED)$$ERRORS remote(s) misconfigured. Fix before continuing.$(RESET)\n"; \
		exit 1; \
	fi
	@echo ""
	@printf "$(BOLD)[3/3] Git hooks$(RESET)\n"
	@printf "  $(INFO) Hooks are Nix-managed via cmdr git module (ADR-005)\n"
	@printf "  $(INFO) Deploy with: unimart deli switch\n"
	@printf "  $(INFO) Per-repo extensions: .githooks/<hook-name>\n"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(GREEN)Bootstrap complete$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

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

status: ## Show submodule state (dirty, ahead/behind, current ref)
ifdef HAS_UNIMART
	@unimart stockroom status
else
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)Submodule Status$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@for mod in $(SUBMODULES); do \
		if [ ! -d "$$mod/.git" ] && [ ! -f "$$mod/.git" ]; then \
			printf "  $(FAIL) %-14s not initialized\n" "$$mod"; \
			continue; \
		fi; \
		\
		BRANCH=$$(git -C "$$mod" rev-parse --abbrev-ref HEAD 2>/dev/null); \
		SHORT=$$(git -C "$$mod" rev-parse --short HEAD 2>/dev/null); \
		TAG=$$(git -C "$$mod" describe --tags --exact-match HEAD 2>/dev/null); \
		\
		DIRTY=""; \
		if [ -n "$$(git -C "$$mod" status --porcelain 2>/dev/null)" ]; then \
			DIRTY=" $(YELLOW)(dirty)$(RESET)"; \
		fi; \
		\
		REF="$$BRANCH@$$SHORT"; \
		if [ -n "$$TAG" ]; then \
			REF="$$TAG ($$BRANCH@$$SHORT)"; \
		fi; \
		\
		printf "  $(CYAN)%-14s$(RESET) %s%b\n" "$$mod" "$$REF" "$$DIRTY"; \
	done
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
endif

update: ## Pull latest main for all submodules
ifdef HAS_UNIMART
	@unimart stockroom update
else
	@printf "$(BOLD)Updating submodules to latest main...$(RESET)\n"
	@echo ""
	@for mod in $(SUBMODULES); do \
		printf "  $(CYAN)$$mod$(RESET) "; \
		BEFORE=$$(git -C "$$mod" rev-parse --short HEAD 2>/dev/null); \
		git -C "$$mod" fetch origin main --quiet 2>/dev/null && \
		git -C "$$mod" checkout main --quiet 2>/dev/null && \
		git -C "$$mod" merge --ff-only origin/main --quiet 2>/dev/null && \
		AFTER=$$(git -C "$$mod" rev-parse --short HEAD 2>/dev/null); \
		if [ "$$BEFORE" = "$$AFTER" ]; then \
			printf "already up to date ($$BEFORE)\n"; \
		else \
			printf "$$BEFORE → $$AFTER\n"; \
		fi; \
	done
	@echo ""
	@printf "$(BOLD)Staging updated submodule pointers...$(RESET)\n"
	@git add $(SUBMODULES)
	@printf "$(GREEN)Done.$(RESET) Review with 'git diff --cached' then commit.\n"
endif

drift: ## Check submodule pointers vs remote HEAD
ifdef HAS_UNIMART
	@unimart stockroom drift
else
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)Drift Check$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@DRIFT=0; \
	for mod in $(SUBMODULES); do \
		if [ ! -d "$$mod/.git" ] && [ ! -f "$$mod/.git" ]; then \
			printf "  $(FAIL) %-14s not initialized\n" "$$mod"; \
			DRIFT=$$((DRIFT + 1)); \
			continue; \
		fi; \
		\
		git -C "$$mod" fetch origin main --quiet 2>/dev/null; \
		LOCAL=$$(git -C "$$mod" rev-parse HEAD 2>/dev/null); \
		REMOTE=$$(git -C "$$mod" rev-parse origin/main 2>/dev/null); \
		\
		if [ "$$LOCAL" = "$$REMOTE" ]; then \
			printf "  $(PASS) %-14s up to date\n" "$$mod"; \
		else \
			BEHIND=$$(git -C "$$mod" rev-list --count HEAD..origin/main 2>/dev/null || echo "?"); \
			AHEAD=$$(git -C "$$mod" rev-list --count origin/main..HEAD 2>/dev/null || echo "?"); \
			STATUS=""; \
			if [ "$$BEHIND" != "0" ] && [ "$$BEHIND" != "?" ]; then \
				STATUS="$$BEHIND behind"; \
			fi; \
			if [ "$$AHEAD" != "0" ] && [ "$$AHEAD" != "?" ]; then \
				if [ -n "$$STATUS" ]; then STATUS="$$STATUS, "; fi; \
				STATUS="$$STATUS$$AHEAD ahead"; \
			fi; \
			printf "  $(WARN) %-14s $$STATUS\n" "$$mod"; \
			DRIFT=$$((DRIFT + 1)); \
		fi; \
	done; \
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ $$DRIFT -eq 0 ]; then \
		printf "$(GREEN)All submodules in sync$(RESET)\n"; \
	else \
		printf "$(YELLOW)$$DRIFT submodule(s) drifted — run 'make update' to sync$(RESET)\n"; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
endif

# ── Cross-Repo ─────────────────────────────────────────────────────────────

ci: ## Validate cross-repo contracts
ifdef HAS_UNIMART
	@unimart stockroom check
else
	@printf "$(BOLD)idpbuilder — Contract Validation$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@ERRORS=0; \
	\
	printf "$(BOLD)[1/6] Submodule initialization$(RESET)\n"; \
	for mod in $(SUBMODULES); do \
		if [ -d "$$mod/.git" ] || [ -f "$$mod/.git" ]; then \
			printf "  $(PASS) $$mod\n"; \
		else \
			printf "  $(FAIL) $$mod not initialized\n"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	done; \
	echo ""; \
	\
	printf "$(BOLD)[2/6] Remote URLs (security check)$(RESET)\n"; \
	for mod in $(SUBMODULES); do \
		REMOTE=$$(git -C "$$mod" remote get-url origin 2>/dev/null); \
		if echo "$$REMOTE" | grep -q 'github.com[:/]idpbuilder/'; then \
			printf "  $(PASS) $$mod → idpbuilder org\n"; \
		else \
			printf "  $(FAIL) $$mod → $$REMOTE (UNEXPECTED REMOTE)\n"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	done; \
	echo ""; \
	\
	printf "$(BOLD)[3/6] AGENTS.md presence$(RESET)\n"; \
	for mod in $(SUBMODULES); do \
		if [ -f "$$mod/AGENTS.md" ]; then \
			printf "  $(PASS) $$mod/AGENTS.md\n"; \
		else \
			printf "  $(WARN) $$mod/AGENTS.md missing\n"; \
		fi; \
	done; \
	if [ -f "AGENTS.md" ]; then \
		printf "  $(PASS) AGENTS.md (org root)\n"; \
	else \
		printf "  $(FAIL) AGENTS.md (org root) missing\n"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	echo ""; \
	\
	printf "$(BOLD)[4/6] Docs directory structure$(RESET)\n"; \
	for mod in cmdr idpbuilder; do \
		if [ -d "$$mod/docs" ]; then \
			MISSING=""; \
			for subdir in Contributing Getting-Started Reference; do \
				if [ ! -d "$$mod/docs/$$subdir" ]; then \
					MISSING="$$MISSING $$subdir"; \
				fi; \
			done; \
			if [ -z "$$MISSING" ]; then \
				printf "  $(PASS) $$mod/docs/ (Contributing, Getting-Started, Reference)\n"; \
			else \
				printf "  $(WARN) $$mod/docs/ missing:$$MISSING\n"; \
			fi; \
		else \
			printf "  $(FAIL) $$mod/docs/ not found\n"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	done; \
	echo ""; \
	\
	printf "$(BOLD)[5/6] Makefile convention$(RESET)\n"; \
	for mod in $(SUBMODULES); do \
		if [ ! -f "$$mod/Makefile" ]; then \
			printf "  $(FAIL) $$mod/Makefile not found\n"; \
			ERRORS=$$((ERRORS + 1)); \
			continue; \
		fi; \
		MISSING_TARGETS=""; \
		for tgt in help hooks; do \
			if ! grep -q "^$$tgt:" "$$mod/Makefile" 2>/dev/null; then \
				MISSING_TARGETS="$$MISSING_TARGETS $$tgt"; \
			fi; \
		done; \
		if [ -z "$$MISSING_TARGETS" ]; then \
			printf "  $(PASS) $$mod — help, hooks\n"; \
		else \
			printf "  $(WARN) $$mod — missing targets:$$MISSING_TARGETS\n"; \
		fi; \
	done; \
	echo ""; \
	\
	printf "$(BOLD)[6/6] Theme export contract$(RESET)\n"; \
	if [ -x "cmdr/scripts/theme-export.sh" ]; then \
		printf "  $(PASS) cmdr/scripts/theme-export.sh (producer)\n"; \
	else \
		printf "  $(FAIL) cmdr/scripts/theme-export.sh not found\n"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	if [ -f "internal/theme/theme.go" ]; then \
		printf "  $(PASS) internal/theme/theme.go (consumer — unimart)\n"; \
	else \
		printf "  $(WARN) internal/theme/theme.go not found\n"; \
	fi; \
	echo ""; \
	\
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ $$ERRORS -eq 0 ]; then \
		printf "$(GREEN)All contract checks passed$(RESET)\n"; \
	else \
		printf "$(RED)$$ERRORS check(s) failed$(RESET)\n"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		exit 1; \
	fi; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
endif

check: ci ## Alias for ci

# ── Testing (container-based) ──────────────────────────────────────────────

# Compose file location
COMPOSE_FILE := containers/podman-compose.yml
COMPOSE_CMD  := podman-compose -f $(COMPOSE_FILE)

test: ## Run make init integration test in Linux container
	@printf "$(BOLD)Building test container...$(RESET)\n"
	@$(COMPOSE_CMD) up -d --build
	@printf "$(BOLD)Running integration test (this takes several minutes)...$(RESET)\n"
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
