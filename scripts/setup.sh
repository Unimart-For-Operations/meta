#!/usr/bin/env bash
# scripts/setup.sh — Unified onboarding: from fresh clone to fully configured environment
#
# Orchestrates the full setup flow:
#   [1] Submodule init
#   [2] Prerequisites (Xcode CLT, Homebrew, Nix)
#   [3] Shell reload (source Nix profile)
#   [4] Host detection / registration
#   [5] Apply configuration
#   [6] Install unimart CLI
#   [7] Verify
#
# Usage:
#   make init                                          # interactive
#   make init GIT_NAME="Alice" GIT_EMAIL="a@b.com"    # non-interactive
#
# Safe to re-run — each step is idempotent.

set -euo pipefail

# ── Colors ────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info() { printf "${CYAN}[info]${RESET}  %s\n" "$*"; }
ok() { printf "${GREEN}[ok]${RESET}    %s\n" "$*"; }
warn() { printf "${YELLOW}[warn]${RESET}  %s\n" "$*"; }
fail() {
	printf "${RED}[fail]${RESET}  %s\n" "$*" >&2
	exit 1
}
step() { printf "\n${BOLD}── %s ──${RESET}\n" "$*"; }
banner() {
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	printf "${BOLD}%s${RESET}\n" "$*"
	echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# ── Resolve repo root ────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CMDR_DIR="${REPO_ROOT}/cmdr"

cd "${REPO_ROOT}"

# ── Accept forwarded parameters ──────────────────────────────────────────
GIT_NAME="${GIT_NAME:-}"
GIT_EMAIL="${GIT_EMAIL:-}"

banner "idpbuilder — Environment Setup"
echo ""

# ── [1/7] Submodules ─────────────────────────────────────────────────────
step "[1/7] Initializing submodules"
if [ -f "${CMDR_DIR}/flake.nix" ]; then
	ok "Submodules already initialized"
else
	info "Running git submodule update --init --recursive..."
	git submodule update --init --recursive
	ok "Submodules initialized"
fi

# ── [2/7] Prerequisites ─────────────────────────────────────────────────
step "[2/7] Installing prerequisites"
if command -v nix &>/dev/null; then
	ok "Nix already installed ($(nix --version))"
	# Still run bootstrap for other prerequisites (Homebrew on macOS)
	if [ "$(uname -s)" = "Darwin" ] && ! command -v brew &>/dev/null; then
		info "Running cmdr bootstrap for Homebrew..."
		bash "${CMDR_DIR}/scripts/bootstrap.sh"
	fi
else
	info "Running cmdr bootstrap (Xcode CLT, Homebrew, Nix)..."
	bash "${CMDR_DIR}/scripts/bootstrap.sh"
fi

# ── [3/7] Shell reload ──────────────────────────────────────────────────
step "[3/7] Ensuring Nix is in PATH"
if ! command -v nix &>/dev/null; then
	# Try sourcing Nix profiles
	if [ -f /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh ]; then
		# shellcheck disable=SC1091
		. /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
	elif [ -f "${HOME}/.nix-profile/etc/profile.d/nix.sh" ]; then
		# shellcheck disable=SC1091
		. "${HOME}/.nix-profile/etc/profile.d/nix.sh"
	fi
fi

if command -v nix &>/dev/null; then
	ok "Nix available in PATH"
else
	warn "Nix not found in current shell PATH"
	echo ""
	echo "  Restart your shell and re-run:"
	echo "    exec zsh"
	echo "    make init"
	exit 0
fi

# ── [4/7] Host detection ────────────────────────────────────────────────
step "[4/7] Detecting host configuration"
CURRENT_USER="$(whoami)"

# Check if a host config already matches this user
HOST_FOUND=""
if [ "$(uname -s)" = "Darwin" ]; then
	SEARCH_DIRS="${CMDR_DIR}/home/02-hosts/macos"
else
	SEARCH_DIRS="${CMDR_DIR}/home/02-hosts/arch ${CMDR_DIR}/home/02-hosts/nixos ${CMDR_DIR}/home/02-hosts/ubuntu"
fi

for search_dir in ${SEARCH_DIRS}; do
	if [ ! -d "${search_dir}" ]; then continue; fi
	for dir in "${search_dir}"/*/; do
		if [ -f "${dir}/meta.nix" ]; then
			username=$(grep 'username.*=' "${dir}/meta.nix" | sed 's/.*username.*=.*"\(.*\)".*/\1/' | tr -d ' ')
			if [ "${username}" = "${CURRENT_USER}" ]; then
				HOST_FOUND="$(basename "${dir}")"
				break 2
			fi
		fi
	done
done

if [ -n "${HOST_FOUND}" ]; then
	ok "Found existing host config: ${HOST_FOUND}"
else
	info "No host config found for user '${CURRENT_USER}'"
	info "Registering this machine..."
	echo ""

	# Build make register args as an array for proper quoting
	REGISTER_ARGS=()
	if [ -n "${GIT_NAME}" ]; then
		REGISTER_ARGS+=("GIT_NAME=${GIT_NAME}")
	fi
	if [ -n "${GIT_EMAIL}" ]; then
		REGISTER_ARGS+=("GIT_EMAIL=${GIT_EMAIL}")
	fi

	# Run register from cmdr directory
	make -C "${CMDR_DIR}" register "${REGISTER_ARGS[@]+"${REGISTER_ARGS[@]}"}"
fi

# ── [5/7] Apply configuration ───────────────────────────────────────────
step "[5/7] Applying configuration"
info "Running make switch in cmdr..."
make -C "${CMDR_DIR}" switch

# ── [6/7] Install unimart CLI ────────────────────────────────────────────
step "[6/7] Installing unimart CLI"
if command -v go &>/dev/null; then
	UNIMART_VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
	UNIMART_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
	UNIMART_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	UNIMART_LDFLAGS="-X github.com/idpbuilder/meta/cmd.Version=${UNIMART_VERSION}"
	UNIMART_LDFLAGS="${UNIMART_LDFLAGS} -X github.com/idpbuilder/meta/cmd.GitCommit=${UNIMART_COMMIT}"
	UNIMART_LDFLAGS="${UNIMART_LDFLAGS} -X github.com/idpbuilder/meta/cmd.BuildDate=${UNIMART_DATE}"

	info "Building unimart (${UNIMART_VERSION})..."
	go build -ldflags "${UNIMART_LDFLAGS}" -o unimart .

	mkdir -p "${HOME}/.local/bin"
	ln -sf "${REPO_ROOT}/unimart" "${HOME}/.local/bin/unimart"

	if "${HOME}/.local/bin/unimart" version &>/dev/null; then
		ok "unimart installed → ~/.local/bin/unimart"
	else
		warn "unimart built but could not verify"
	fi

	# Check PATH includes ~/.local/bin
	case ":${PATH}:" in
	*":${HOME}/.local/bin:"*) ;;
	*) warn "~/.local/bin is not in PATH — add it to your shell profile" ;;
	esac
else
	warn "Go not found — skipping unimart install"
	info "Run 'make install' after reloading your shell"
fi

# ── [7/7] Verify ─────────────────────────────────────────────────────────
step "[7/7] Verifying environment"
make -C "${CMDR_DIR}" doctor

# ── Done ──────────────────────────────────────────────────────────────────
echo ""
banner "Setup complete"
echo ""
echo "  Reload your shell to activate the new configuration:"
echo "    exec zsh"
echo ""
echo "  Useful commands:"
echo "    unimart deli doctor       # Check system health"
echo "    unimart deli hosts        # List host configs"
echo "    unimart stockroom drift   # Check submodule drift"
echo "    unimart --help            # Browse all aisles"
echo ""
echo "  Or use make targets:"
echo "    make help                 # Show all available targets"
echo ""
