#!/usr/bin/env bash
# containers/test-init.sh — Integration test for the full `make init` onboarding flow
#
# Runs inside the Ubuntu 24.04 test container. Expects:
#   - Meta repo mounted read-only at /workspace
#   - Nix either installed or install-nix.sh available
#
# Usage (called by Makefile, not directly):
#   podman-compose exec -T init-test bash /workspace/containers/test-init.sh

set -euo pipefail

# ── Colors ────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info() { printf "${CYAN}[info]${RESET}  %s\n" "$*"; }
ok() { printf "${GREEN}[pass]${RESET}  %s\n" "$*"; }
# shellcheck disable=SC2329  # kept for symmetry with info/ok/fail_msg
warn() { printf "${YELLOW}[warn]${RESET}  %s\n" "$*"; }
fail_msg() { printf "${RED}[FAIL]${RESET}  %s\n" "$*"; }

banner() {
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	printf "${BOLD}%s${RESET}\n" "$*"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# ── Test identity ─────────────────────────────────────────────────────────
TEST_GIT_NAME="Test User"
TEST_GIT_EMAIL="test@example.com"

# ── Workspace ─────────────────────────────────────────────────────────────
WRITABLE_DIR="/home/nixuser/idpbuilder"

banner "idpbuilder — Integration Test: make init"
echo ""

# ── [1/6] Install Nix ────────────────────────────────────────────────────
info "[1/6] Ensuring Nix is installed"
if command -v nix &>/dev/null; then
	ok "Nix already installed ($(nix --version))"
elif [ -f /nix/var/nix/profiles/default/bin/nix ]; then
	ok "Nix binary found, sourcing profile"
	# shellcheck disable=SC1091
	. /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
else
	info "Installing Nix..."
	/home/nixuser/install-nix.sh
	# shellcheck disable=SC1091
	. /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
fi

# ── [2/6] Start nix-daemon ───────────────────────────────────────────────
info "[2/6] Starting nix-daemon"
if pgrep -x nix-daemon &>/dev/null; then
	ok "nix-daemon already running"
else
	sudo /nix/var/nix/profiles/default/bin/nix-daemon </dev/null 2>/tmp/nix-daemon.log &
	sleep 2
	if pgrep -x nix-daemon &>/dev/null; then
		ok "nix-daemon started"
	else
		fail_msg "Failed to start nix-daemon"
		exit 1
	fi
fi

# ── [3/6] Copy repo to writable location ────────────────────────────────
info "[3/6] Copying meta repo to writable location"
if [ -d "${WRITABLE_DIR}" ]; then
	info "Cleaning previous copy..."
	rm -rf "${WRITABLE_DIR}"
fi
cp -a /workspace "${WRITABLE_DIR}"

# Fix git safe.directory for the copy
git config --global --add safe.directory "${WRITABLE_DIR}"
git config --global --add safe.directory "${WRITABLE_DIR}/cmdr"
git config --global --add safe.directory "${WRITABLE_DIR}/idpbuilder"
git config --global --add safe.directory "${WRITABLE_DIR}/docs"

# Remove any pre-existing .zshrc that conflicts with Home Manager
rm -f /home/nixuser/.zshrc

ok "Repo copied to ${WRITABLE_DIR}"

# ── [4/6] Run make init ─────────────────────────────────────────────────
info "[4/6] Running make init with test identity"
echo ""

INIT_EXIT=0
GIT_NAME="${TEST_GIT_NAME}" GIT_EMAIL="${TEST_GIT_EMAIL}" \
	make -C "${WRITABLE_DIR}" init || INIT_EXIT=$?

echo ""

if [ "${INIT_EXIT}" -eq 0 ]; then
	ok "make init completed (exit code 0)"
else
	fail_msg "make init failed (exit code ${INIT_EXIT})"
fi

# ── [5/6] Assertions ────────────────────────────────────────────────────
banner "Assertions"
echo ""

FAILURES=0

# Assert: meta.nix was created for this host
info "Checking host registration..."
HOST_META=$(find "${WRITABLE_DIR}/cmdr/home/02-hosts" -name "meta.nix" -path "*$(hostname)*" 2>/dev/null || true)
if [ -n "${HOST_META}" ]; then
	ok "Host meta.nix found: ${HOST_META}"
else
	# Also check by username since hostname may not match directory name
	HOST_META=$(find "${WRITABLE_DIR}/cmdr/home/02-hosts" -name "meta.nix" -exec grep -l "$(whoami)" {} \; 2>/dev/null | head -1 || true)
	if [ -n "${HOST_META}" ]; then
		ok "Host meta.nix found (by username): ${HOST_META}"
	else
		fail_msg "No meta.nix found for this host or user"
		FAILURES=$((FAILURES + 1))
	fi
fi

# Assert: git identity is configured
info "Checking git identity..."
ACTUAL_NAME=$(git config --global user.name 2>/dev/null || true)
ACTUAL_EMAIL=$(git config --global user.email 2>/dev/null || true)

if [ "${ACTUAL_NAME}" = "${TEST_GIT_NAME}" ]; then
	ok "git user.name = '${ACTUAL_NAME}'"
else
	fail_msg "git user.name = '${ACTUAL_NAME}' (expected '${TEST_GIT_NAME}')"
	FAILURES=$((FAILURES + 1))
fi

if [ "${ACTUAL_EMAIL}" = "${TEST_GIT_EMAIL}" ]; then
	ok "git user.email = '${ACTUAL_EMAIL}'"
else
	fail_msg "git user.email = '${ACTUAL_EMAIL}' (expected '${TEST_GIT_EMAIL}')"
	FAILURES=$((FAILURES + 1))
fi

# Assert: key binaries on PATH
info "Checking key binaries..."
for bin in git zsh nix; do
	if command -v "${bin}" &>/dev/null; then
		ok "${bin} on PATH ($(command -v "${bin}"))"
	else
		fail_msg "${bin} not found on PATH"
		FAILURES=$((FAILURES + 1))
	fi
done

# Assert: Home Manager generation exists
info "Checking Home Manager..."
HM_GENS=$(home-manager generations 2>/dev/null | head -1 || true)
if [ -n "${HM_GENS}" ]; then
	ok "Home Manager generation exists: ${HM_GENS}"
else
	fail_msg "No Home Manager generations found"
	FAILURES=$((FAILURES + 1))
fi

# ── [6/6] Results ────────────────────────────────────────────────────────
echo ""
banner "Test Results"
echo ""

TOTAL_CHECKS=6 # meta.nix, git name, git email, git, zsh, nix, HM generation
PASSED=$((TOTAL_CHECKS - FAILURES))

if [ "${INIT_EXIT}" -ne 0 ]; then
	fail_msg "make init itself failed — assertions may be incomplete"
	FAILURES=$((FAILURES + 1))
fi

if [ "${FAILURES}" -eq 0 ]; then
	printf "${GREEN}${BOLD}ALL CHECKS PASSED${RESET} (%d/%d)\n" "${PASSED}" "${TOTAL_CHECKS}"
	echo ""
	exit 0
else
	printf "${RED}${BOLD}%d FAILURE(S)${RESET} (%d/%d passed)\n" "${FAILURES}" "${PASSED}" "${TOTAL_CHECKS}"
	echo ""
	exit 1
fi
