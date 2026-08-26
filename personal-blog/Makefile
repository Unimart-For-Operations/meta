.PHONY: help server test fmt lint ci hooks sync-docs
.DEFAULT_GOAL := help

CYAN   := \033[0;36m
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
RESET  := \033[0m
BOLD   := \033[1m

PASS := \033[0;32m[pass]\033[0m
FAIL := \033[0;31m[fail]\033[0m
WARN := \033[0;33m[warn]\033[0m

ORG_DIR := $(shell dirname "$(CURDIR)")

help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@printf "$(BOLD)personal-blog — Phoenix microservice for portfolio, projects, articles$(RESET)\n"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@printf "$(BOLD)Development$(RESET)\n"
	@printf "  $(CYAN)server$(RESET)         Start development server (localhost:4000)\n"
	@printf "  $(CYAN)test$(RESET)           Run test suite\n"
	@printf "  $(CYAN)fmt$(RESET)            Format Elixir code\n"
	@printf "  $(CYAN)lint$(RESET)           Lint code with Credo\n"
	@printf "  $(CYAN)setup$(RESET)          Install dependencies and setup\n"
	@echo ""
	@printf "$(BOLD)CI/CD$(RESET)\n"
	@printf "  $(CYAN)ci$(RESET)             Run all local checks\n"
	@printf "  $(CYAN)build$(RESET)          Build Docker image\n"
	@echo ""
	@printf "$(BOLD)Platform$(RESET)\n"
	@printf "  $(CYAN)sync-docs$(RESET)      Sync docs to handbook\n"
	@printf "  $(CYAN)hooks$(RESET)          Install git hooks\n"
	@echo ""

## Development targets

setup:
	@printf "\n$(BOLD)[1/2] Installing dependencies$(RESET)\n"
	@mix local.hex --if-missing && mix local.rebar --if-missing && mix deps.get
	@printf "  $(PASS) Dependencies installed\n"
	@printf "\n$(BOLD)[2/2] Setting up assets$(RESET)\n"
	@mix tailwind.install --if-missing
	@printf "  $(PASS) Assets configured\n"

server:
	@mix phx.server

test:
	@mix test

fmt:
	@mix format

lint:
	@mix credo suggest --strict

## CI targets

ci:
	@ERRORS=0; \
	printf "\n$(BOLD)[1/4] Formatting check$(RESET)\n"; \
	if mix format --check-formatted >/dev/null 2>&1; then printf "  $(PASS) formatted correctly\n"; \
	else printf "  $(FAIL) formatting issues found\n"; ERRORS=$$((ERRORS + 1)); fi; \
	\
	printf "\n$(BOLD)[2/4] Linting$(RESET)\n"; \
	if mix credo suggest --strict 2>/dev/null | grep -q "0 issues"; then printf "  $(PASS) no lint issues\n"; \
	else printf "  $(WARN) see credo output above\n"; fi; \
	\
	printf "\n$(BOLD)[3/4] Compilation$(RESET)\n"; \
	if mix compile 2>&1 | grep -q -i error; then printf "  $(FAIL) compilation errors\n"; ERRORS=$$((ERRORS + 1)); \
	else printf "  $(PASS) compiles successfully\n"; fi; \
	\
	printf "\n$(BOLD)[4/4] Tests$(RESET)\n"; \
	if mix test; then printf "  $(PASS) all tests pass\n"; \
	else printf "  $(FAIL) test failures\n"; ERRORS=$$((ERRORS + 1)); fi; \
	\
	echo ""; \
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	if [ $$ERRORS -eq 0 ]; then \
	  printf "$(GREEN)All checks passed$(RESET)\n"; \
	else \
	  printf "$(RED)$$ERRORS check(s) failed$(RESET)\n"; exit 1; \
	fi

build:
	@printf "\n$(BOLD)Building Docker image$(RESET)\n"
	@docker build -t personal-blog:latest .
	@printf "$(GREEN)✓ Image built$(RESET)\n"

## Platform targets

sync-docs:
	@if command -v unimart >/dev/null 2>&1; then \
	  unimart newsstand sync; \
	elif [ -x "$(ORG_DIR)/unimart-employee-handbooks/cdc/scripts/sync-docs.sh" ]; then \
	  bash "$(ORG_DIR)/unimart-employee-handbooks/cdc/scripts/sync-docs.sh"; \
	else \
	  printf "$(RED)Docs sync unavailable$(RESET)\n"; exit 1; \
	fi

hooks:
	@if command -v unimart >/dev/null 2>&1; then \
	  printf "$(CYAN)Installing git hooks via unimart...$(RESET)\n"; \
	  unimart deli switch; \
	else \
	  printf "$(YELLOW)unimart not available; git hooks not installed$(RESET)\n"; \
	fi
