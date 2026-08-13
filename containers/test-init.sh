#!/usr/bin/env bash
# containers/test-init.sh — CLI-first smoke test for the unimart workflow
#
# Runs inside the Ubuntu 24.04 test container. Expects:
#   - Meta repo mounted read-only at /workspace
#   - Go toolchain available
#   - Nix either installed or install-nix.sh available for the CLI bootstrap path
#
# Usage (called by Makefile, not directly):
#   docker compose exec -T init-test bash /workspace/containers/test-init.sh

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info() { printf "${CYAN}[info]${RESET}  %s\n" "$*"; }
ok() { printf "${GREEN}[pass]${RESET}  %s\n" "$*"; }
warn() { printf "${YELLOW}[warn]${RESET}  %s\n" "$*"; }
fail_msg() { printf "${RED}[FAIL]${RESET}  %s\n" "$*"; }

banner() {
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	printf "${BOLD}%s${RESET}\n" "$*"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

WRITABLE_DIR="/home/nixuser/meta-cli-test"

banner "meta — CLI-first smoke test"
echo ""

info "[1/5] Ensuring Go toolchain is available"
if command -v go &>/dev/null; then
	ok "Go available ($(go version 2>/dev/null || echo unknown))"
else
	fail_msg "Go not found on PATH"
	exit 1
fi

info "[2/5] Copying repo to writable location"
if [ -d "${WRITABLE_DIR}" ]; then
	rm -rf "${WRITABLE_DIR}"
fi
mkdir -p "${WRITABLE_DIR}"

# Rootless podman userns maps host uid 1000 to root inside the container, so
# host-private files (mode 600/700) appear root-owned and unreadable by nixuser.
# tar --ignore-failed-read skips those without failing the whole copy.
tar -C /workspace -cf - . --ignore-failed-read | tar -xf - -C "${WRITABLE_DIR}"

git config --global --add safe.directory "${WRITABLE_DIR}"

authors=$(git config --global --get-regexp 'safe\.directory' || true)
if [ -n "${authors}" ]; then
	ok "Safe.directory configured for test repo"
fi

info "[3/5] Building unimart from source"
cd "${WRITABLE_DIR}"
if go build -o /tmp/unimart .; then
	ok "unimart builds successfully"
else
	fail_msg "unimart build failed"
	exit 1
fi

info "[4/5] Running CLI health checks"
CLI_EXIT=0
/tmp/unimart deli doctor || CLI_EXIT=$?
if [ "${CLI_EXIT}" -eq 0 ]; then
	ok "deli doctor completed cleanly"
else
	warn "deli doctor reported issues (non-fatal for smoke test)"
fi

info "[5/5] Running contract validation"
CHECK_EXIT=0
/tmp/unimart stockroom check || CHECK_EXIT=$?
if [ "${CHECK_EXIT}" -eq 0 ]; then
	ok "stockroom check passed"
else
	fail_msg "stockroom check failed (exit ${CHECK_EXIT})"
	exit 1
fi

echo ""
banner "Test Results"
echo ""
printf "${GREEN}${BOLD}CLI smoke test passed${RESET}\n"
echo ""
exit 0
