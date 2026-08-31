#!/usr/bin/env bash
# scripts/provision.sh — One-line bootstrap for fresh machines
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Unimart-For-Operations/meta/main/scripts/provision.sh | bash
#
# Or with custom clone location:
#   META_DIR=~/my-org bash <(curl -fsSL ...)
#
# Prerequisites (platform-specific):
#   macOS:   (none - bash + curl pre-installed)
#   Arch:    sudo pacman -S git curl
#   NixOS:   (usually pre-installed)
#   Ubuntu:  sudo apt install git curl

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────
REPO_URL="https://github.com/Unimart-For-Operations/meta.git"
REPO_BRANCH="main"
CLONE_DIR="${META_DIR:-${HOME}/repos/Unimart-For-Operations/meta}"

# ── Colors ────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────
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

# ── Platform Detection ────────────────────────────────────────────────
detect_platform() {
	OS="$(uname -s)"
	case "$OS" in
	Darwin) echo "macOS" ;;
	Linux)
		if [[ -f /etc/NIXOS ]]; then
			echo "NixOS"
		elif [[ -f /etc/os-release ]]; then
			. /etc/os-release
			echo "$ID"
		else
			echo "Linux"
		fi
		;;
	*) echo "unknown" ;;
	esac
}

PLATFORM=$(detect_platform)

banner "Unimart-For-Operations — One-Line Provisioning"
info "Detected platform: $PLATFORM"
echo ""

# ── [1/6] Pre-flight Checks ───────────────────────────────────────────
step "[1/6] Pre-flight checks"

if ! command -v git &>/dev/null; then
	echo ""
	fail "git not found. Install it first:\n\n  macOS:  xcode-select --install\n  Arch:   sudo pacman -S git\n  NixOS:  nix-env -iA nixos.git\n  Ubuntu: sudo apt install git"
fi
ok "git installed"

if ! command -v curl &>/dev/null; then
	echo ""
	fail "curl not found. Install it first:\n\n  macOS:  (pre-installed)\n  Arch:   sudo pacman -S curl\n  NixOS:  nix-env -iA nixos.curl\n  Ubuntu: sudo apt install curl"
fi
ok "curl installed"

# ── [2/6] Clone Repository ────────────────────────────────────────────
step "[2/6] Cloning repository"

if [[ -d "$CLONE_DIR/.git" ]]; then
	ok "Repository already exists at $CLONE_DIR"
	cd "$CLONE_DIR"
	# Pull latest changes
	info "Pulling latest changes..."
	git pull origin "$REPO_BRANCH" 2>/dev/null || warn "Could not pull latest changes (non-fatal)"
else
	info "Cloning $REPO_URL to $CLONE_DIR..."
	mkdir -p "$(dirname "$CLONE_DIR")"
	git clone --branch "$REPO_BRANCH" "$REPO_URL" "$CLONE_DIR"
	ok "Repository cloned"
	cd "$CLONE_DIR"
fi

# ── [3/6] Install Prerequisites ───────────────────────────────────────
step "[3/6] Installing prerequisites"

BOOTSTRAP_SCRIPT="$CLONE_DIR/cmdr/scripts/bootstrap.sh"
if [[ ! -f "$BOOTSTRAP_SCRIPT" ]]; then
	fail "Bootstrap script not found at $BOOTSTRAP_SCRIPT"
fi

bash "$BOOTSTRAP_SCRIPT"

# ── [4/6] Verify Nix ──────────────────────────────────────────────────
step "[4/6] Verifying Nix installation"

# Source Nix profile if not already in PATH
if ! command -v nix &>/dev/null; then
	info "Nix not in PATH, attempting to source profile..."
	# shellcheck disable=SC1091
	if [[ -f /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh ]]; then
		. /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
	elif [[ -f "$HOME/.nix-profile/etc/profile.d/nix.sh" ]]; then
		. "$HOME/.nix-profile/etc/profile.d/nix.sh"
	elif [[ -f /etc/profile.d/nix.sh ]]; then
		. /etc/profile.d/nix.sh
	fi
fi

if command -v nix &>/dev/null; then
	NIX_VERSION=$(nix --version)
	ok "Nix available ($NIX_VERSION)"
else
	warn "Nix installed but not available in current shell"
	echo ""
	echo "  Restart your shell and re-run this script:"
	echo "    exec zsh"
	echo "    curl -fsSL https://raw.githubusercontent.com/Unimart-For-Operations/meta/main/scripts/provision.sh | bash"
	exit 0
fi

# ── [5/6] Initialize Submodules ───────────────────────────────────────
step "[5/6] Initializing submodules"

if [[ -f "$CLONE_DIR/cmdr/flake.nix" ]]; then
	ok "Submodules already initialized"
else
	info "Running git submodule update --init --recursive..."
	git submodule update --init --recursive
	ok "Submodules initialized"
fi

# ── [6/6] Bootstrap via unimart ───────────────────────────────────────
step "[6/6] Running unimart bootstrap"
echo ""

info "Building and running unimart from flake..."
echo ""

# Run unimart bootstrap from the flake (no local Go required)
nix run "$CLONE_DIR#unimart" -- deli bootstrap

# ── Done ──────────────────────────────────────────────────────────────
echo ""
banner "Provisioning Complete"
echo ""
echo "  Next steps:"
echo ""
echo "    1. Reload your shell to activate the new configuration:"
echo "         exec zsh"
echo ""
echo "    2. If a new host was created, commit and push it:"
echo "         cd $CLONE_DIR"
echo "         git add cmdr/home/02-hosts/"
echo "         git commit -s -m \"feat(host): add \$(hostname) configuration\""
echo "         git push origin main"
echo ""
echo "    3. Verify everything is working:"
echo "         unimart deli doctor"
echo ""
echo "  Useful commands:"
echo "    unimart deli hosts        # List all host configurations"
echo "    unimart deli switch       # Apply configuration changes"
echo "    unimart freezer up        # Start local IDP platform"
echo "    unimart --help            # Browse all commands"
echo ""
