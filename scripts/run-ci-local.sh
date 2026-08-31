#!/usr/bin/env bash
#
# run-ci-local.sh - Run CI checks locally before pushing
#
# Usage:
#   ./scripts/run-ci-local.sh           # Run all checks
#   ./scripts/run-ci-local.sh shellcheck # Run specific check
#
# Exit codes:
#   0 - All checks passed
#   1 - One or more checks failed

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Change to repo root
cd "$(git rev-parse --show-toplevel)"

# Track failures
FAILED_CHECKS=()

# Helper functions
print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

pass() {
    echo -e "${GREEN}✅ $1${NC}"
}

fail() {
    echo -e "${RED}❌ $1${NC}"
    FAILED_CHECKS+=("$1")
}

warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check functions
check_shellcheck() {
    print_header "ShellCheck (Scripts)"
    
    if ! command -v shellcheck &> /dev/null; then
        warn "shellcheck not installed, skipping"
        return 0
    fi
    
    echo "Checking meta shell scripts..."
    if find scripts/ -type f -name "*.sh" -print0 | xargs -0 shellcheck; then
        pass "Shell scripts passed shellcheck"
    else
        fail "ShellCheck failed"
        return 1
    fi
}

check_go() {
    print_header "Go (fmt, vet, build, test)"
    
    local FAILED=0
    
    echo "Running go fmt (meta)..."
    if ! go fmt ./...; then
        fail "go fmt (meta) failed"
        FAILED=1
    fi
    
    echo "Running go fmt (idpbuilder)..."
    if ! go -C idpbuilder fmt ./...; then
        fail "go fmt (idpbuilder) failed"
        FAILED=1
    fi
    
    echo "Running go vet (meta)..."
    if ! go vet ./...; then
        fail "go vet (meta) failed"
        FAILED=1
    fi
    
    echo "Running go vet (idpbuilder)..."
    if ! go -C idpbuilder vet ./...; then
        fail "go vet (idpbuilder) failed"
        FAILED=1
    fi
    
    echo "Building unimart..."
    if ! go build -o unimart .; then
        fail "go build failed"
        FAILED=1
    fi
    
    echo "Running go test (meta)..."
    if ! go test ./...; then
        fail "go test (meta) failed"
        FAILED=1
    fi
    
    echo "Running go test (idpbuilder)..."
    if ! go -C idpbuilder test ./...; then
        fail "go test (idpbuilder) failed"
        FAILED=1
    fi
    
    if [ $FAILED -eq 0 ]; then
        pass "All Go checks passed"
    else
        return 1
    fi
}

check_nix() {
    print_header "Nix (flake check)"
    
    if ! command -v nix &> /dev/null; then
        warn "nix not installed, skipping"
        return 0
    fi
    
    echo "Initializing cmdr submodule..."
    git submodule update --init cmdr
    
    echo "Running nix flake check (cmdr)..."
    if ! (cd cmdr && nix flake check --all-systems); then
        fail "nix flake check failed"
        return 1
    fi
    
    echo "Running nix fmt check (cmdr)..."
    if ! (cd cmdr && nix fmt -- --check .); then
        fail "nix fmt check failed"
        return 1
    fi
    
    pass "Nix checks passed"
}

check_docs() {
    print_header "Documentation (markdown, paths)"
    
    local FAILED=0
    
    # Markdown lint (non-fatal)
    if command -v markdownlint &> /dev/null; then
        echo "Running markdown lint..."
        if ! markdownlint '**/*.md' --ignore node_modules --ignore .git --config .markdownlint.json; then
            warn "Markdown lint found issues (non-fatal)"
        fi
    else
        warn "markdownlint not installed, skipping"
    fi
    
    # Path consistency check
    echo "Checking for old path references..."
    
    EXCLUDE_DOCS=(
        "CI.md"
        "CI-TEST-RESULTS.md"
        "onboarding-improvements-2026-08-31.md"
        "COMMIT-CHECKLIST.md"
        "SUMMARY.md"
        "run-ci-local.sh"
    )
    
    GREP_EXCLUDES=""
    for doc in "${EXCLUDE_DOCS[@]}"; do
        GREP_EXCLUDES="$GREP_EXCLUDES --exclude=$doc"
    done
    
    # shellcheck disable=SC2086
    if grep -r "repos/github/idpbuilder" \
        --include="*.md" \
        --include="*.sh" \
        --include="*.nix" \
        --exclude-dir=.git \
        --exclude-dir=node_modules \
        --exclude-dir=unimart-employee-handbooks \
        $GREP_EXCLUDES \
        . 2>/dev/null; then
        fail "Found old path references"
        FAILED=1
    fi
    
    # Check provision.sh
    echo "Checking provision.sh..."
    if [ ! -f scripts/provision.sh ]; then
        fail "scripts/provision.sh not found"
        FAILED=1
    elif [ ! -x scripts/provision.sh ]; then
        fail "scripts/provision.sh not executable"
        FAILED=1
    fi
    
    if [ $FAILED -eq 0 ]; then
        pass "Documentation checks passed"
    else
        return 1
    fi
}

check_hosts() {
    print_header "Host Configs (NixOS)"
    
    if ! command -v nix &> /dev/null; then
        warn "nix not installed, skipping"
        return 0
    fi
    
    echo "Initializing cmdr submodule..."
    git submodule update --init cmdr
    
    cd cmdr
    
    local FAILED=0
    
    for host_dir in home/02-hosts/nixos/*/; do
        host_name=$(basename "$host_dir")
        
        # Skip template
        if [ "$host_name" = "_template" ]; then
            continue
        fi
        
        echo "Checking $host_name..."
        
        # Check required files
        for file in meta.nix default.nix system.nix hardware-configuration.nix; do
            if [ ! -f "${host_dir}${file}" ]; then
                fail "$host_name: Missing ${file}"
                FAILED=1
            fi
        done
        
        # Evaluate configuration
        if ! nix eval --json ".#nixosConfigurations.${host_name}.config.system.build.toplevel" > /dev/null 2>&1; then
            fail "$host_name: Configuration evaluation failed"
            FAILED=1
        else
            echo "  ✓ $host_name configuration is valid"
        fi
    done
    
    cd ..
    
    if [ $FAILED -eq 0 ]; then
        pass "All host configs valid"
    else
        return 1
    fi
}

check_provision() {
    print_header "Provision Script (smoke test)"
    
    echo "Testing provision.sh syntax..."
    if ! bash -n scripts/provision.sh; then
        fail "provision.sh has syntax errors"
        return 1
    fi
    
    pass "Provision script syntax valid"
}

# Main execution
main() {
    local CHECK="${1:-all}"
    
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Local CI Checks - Meta Repository${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    
    case "$CHECK" in
        all)
            check_shellcheck || true
            check_go || true
            check_nix || true
            check_docs || true
            check_hosts || true
            check_provision || true
            ;;
        shellcheck)
            check_shellcheck
            ;;
        go)
            check_go
            ;;
        nix)
            check_nix
            ;;
        docs)
            check_docs
            ;;
        hosts)
            check_hosts
            ;;
        provision)
            check_provision
            ;;
        *)
            echo -e "${RED}Unknown check: $CHECK${NC}"
            echo "Available checks: all, shellcheck, go, nix, docs, hosts, provision"
            exit 1
            ;;
    esac
    
    # Summary
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    if [ ${#FAILED_CHECKS[@]} -eq 0 ]; then
        echo -e "${GREEN}✅ All checks passed!${NC}"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        exit 0
    else
        echo -e "${RED}❌ ${#FAILED_CHECKS[@]} check(s) failed:${NC}"
        for check in "${FAILED_CHECKS[@]}"; do
            echo -e "${RED}  - $check${NC}"
        done
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        exit 1
    fi
}

main "$@"
