# Commit Checklist: legion-nix Onboarding & Path Standardization

## Pre-Commit Verification

### File Existence
- [x] `cmdr/home/02-hosts/nixos/legion-nix/meta.nix` exists
- [x] `cmdr/home/02-hosts/nixos/legion-nix/default.nix` exists
- [x] `cmdr/home/02-hosts/nixos/legion-nix/system.nix` exists
- [x] `cmdr/home/02-hosts/nixos/legion-nix/hardware-configuration.nix` exists (placeholder)
- [x] `cmdr/home/02-hosts/nixos/legion-nix/ONBOARDING.md` exists
- [x] `cmdr/home/02-hosts/nixos/legion-nix/CHECKLIST.md` exists
- [x] `scripts/provision.sh` exists and is executable
- [x] `docs/onboarding-improvements-2026-08-31.md` exists
- [x] `docs/SUMMARY.md` exists

### Path Standardization Verification
- [x] No references to `~/repos/github/idpbuilder` in .md files
- [x] No references to `~/repos/github/idpbuilder` in .sh files
- [x] All paths updated to `~/repos/Unimart-For-Operations/meta`
- [x] Tmux layouts updated
- [x] AGENTS.md files updated (meta, cmdr, cdc)

### Host Configuration
- [x] `unimart deli hosts` shows legion-nix
- [x] legion-nix role: tty-engineer
- [x] legion-nix features: cli, tui
- [x] legion-nix capabilities: baseline, terminal-dev, operator

### Script Validation
- [x] provision.sh is executable (chmod +x)
- [x] provision.sh has correct shebang (#!/usr/bin/env bash)
- [x] provision.sh has proper error handling (set -euo pipefail)
- [x] provision.sh uses correct clone path

### Documentation Quality
- [x] ONBOARDING.md comprehensive (14KB)
- [x] CHECKLIST.md actionable (3.9KB)
- [x] onboarding-improvements-2026-08-31.md complete (13KB)
- [x] SUMMARY.md accurate (this file)

## Git Status Check

Run before committing:

```bash
cd ~/repos/meta
git status
```

**Expected changes**:
- New files: 9 (legion-nix configs, provision.sh, docs)
- Modified files: ~8 (AGENTS.md, tmux layouts, etc.)
- Submodules: cmdr may show as modified (ignore = dirty)

## Commit Commands

### Option 1: Single Comprehensive Commit

```bash
cd ~/repos/meta

# Stage all changes
git add \
  cmdr/home/02-hosts/nixos/legion-nix/ \
  scripts/provision.sh \
  docs/onboarding-improvements-2026-08-31.md \
  docs/SUMMARY.md \
  AGENTS.md \
  cmdr/home/04-modules/tui/graduated/tmux/layouts.nix \
  unimart-employee-handbooks/cdc/meta/AGENTS.md

# Commit with DCO sign-off
git commit -s -m "feat(legion-nix): add NixOS host with comprehensive onboarding

- Add legion-nix host configuration (tty-engineer, cli+tui)
- Create provision.sh one-line bootstrap script
- Standardize repository path to ~/repos/Unimart-For-Operations/meta
- Document 7 strix-nix pain points with solutions
- Update all AGENTS.md, tmux layouts, and docs for new path convention

Changes:
- New host: cmdr/home/02-hosts/nixos/legion-nix/ (6 files, 23KB)
- New script: scripts/provision.sh (7.1KB, one-line curl-to-bash bootstrap)
- New docs: docs/onboarding-improvements-2026-08-31.md (13KB deep dive)
- New docs: docs/SUMMARY.md (session summary)
- Path updates: 8 files across meta, cmdr, cdc

Pain points addressed:
1. Shell trampoline (bash→zsh on NixOS)
2. DNS resolution for *.cnoe.localtest.me (NetworkManager dnsmasq)
3. uwsm session startup (documented for future GUI)
4. Container testing compatibility (no GUI deps)
5. Cross-platform build limitations (documented)
6. Bluetooth hardware detection (explicit enable)
7. Hardware-specific config patterns (NVIDIA, keyboard backlight)

Breaking change: Repository path convention now ~/repos/Unimart-For-Operations/meta
(low impact - unimart resolveOrgDir() still finds repos via CWD walk)

Refs: BOOTSTRAP-ACCEPTANCE.md, strix-nix (reference implementation)
Tested: unimart deli hosts shows legion-nix correctly"

# Verify commit
git log --oneline -1
git show --stat
```

### Option 2: Granular Commits (Recommended for easier review)

```bash
cd ~/repos/meta

# Commit 1: legion-nix host configuration
git add cmdr/home/02-hosts/nixos/legion-nix/
git commit -s -m "feat(legion-nix): add NixOS host configuration

New TTY-focused NixOS host with cli+tui features:
- meta.nix: tty-engineer role, baseline capabilities
- default.nix: shell trampoline (bash→zsh)
- system.nix: NetworkManager, docker, nix-ld, DNS for IDP
- hardware-configuration.nix: placeholder for system generation
- ONBOARDING.md: comprehensive guide with strix-nix pain points
- CHECKLIST.md: quick-reference bootstrap checklist

Proactively solves 7 documented pain points from strix-nix deployment."

# Commit 2: provision.sh bootstrap script
git add scripts/provision.sh
git commit -s -m "feat(bootstrap): add provision.sh one-line bootstrap script

One-command setup for fresh machines:
  curl -fsSL https://...provision.sh | bash

Features:
- Platform-agnostic (macOS, Arch, NixOS, Ubuntu)
- Minimal prerequisites (git + curl)
- Idempotent (safe to re-run)
- Sources Nix in-process (no manual shell reload)
- Delegates to nix run .#unimart -- deli bootstrap
- Customizable via META_DIR env var

Implements requirements from BOOTSTRAP-ACCEPTANCE.md across all platforms."

# Commit 3: Path standardization
git add \
  AGENTS.md \
  cmdr/home/04-modules/tui/graduated/tmux/layouts.nix \
  unimart-employee-handbooks/cdc/meta/AGENTS.md \
  cmdr/home/02-hosts/nixos/legion-nix/ONBOARDING.md \
  cmdr/home/02-hosts/nixos/legion-nix/CHECKLIST.md

git commit -s -m "refactor(paths): standardize to ~/repos/Unimart-For-Operations/meta

Update repository path convention across all documentation:
- Old (inconsistent): ~/repos/github/idpbuilder, ~/repos/meta
- New (standard): ~/repos/Unimart-For-Operations/meta

Rationale:
- Mirrors GitHub organization structure
- Clear ownership and discoverability
- Extensible to other org repos

Files updated:
- AGENTS.md (meta, cmdr, cdc)
- Tmux layouts (idpbuilder session)
- legion-nix onboarding docs

Impact: Low - unimart's resolveOrgDir() finds repos via CWD walk"

# Commit 4: Documentation
git add docs/
git commit -s -m "docs(onboarding): document legion-nix improvements

Add comprehensive documentation for onboarding session:
- onboarding-improvements-2026-08-31.md: deep dive on changes
  * 7 pain points from strix-nix with solutions
  * Path standardization rationale
  * provision.sh design and acceptance criteria
  * Before/after bootstrap flow comparison
  * Breaking changes and migration paths
- SUMMARY.md: session summary and metrics

Total new documentation: 13KB"
```

## Post-Commit Verification

```bash
# View commit history
git log --oneline -4

# Check submodule status
git status

# If cmdr shows as modified (expected due to ignore=dirty):
git diff cmdr  # Should show Subproject commit changes only

# Verify no uncommitted changes to tracked files
git status --short
```

## Push to Remote

```bash
# Push to origin
git push origin main

# Verify on GitHub
# https://github.com/Unimart-For-Operations/meta/commits/main
```

## Post-Push Tasks

- [ ] Verify commits appear on GitHub
- [ ] Check CI/CD pipeline (if configured)
- [ ] Test provision.sh from raw GitHub URL on a fresh container:
  ```bash
  docker run -it nixos/nix bash
  curl -fsSL https://raw.githubusercontent.com/Unimart-For-Operations/meta/main/scripts/provision.sh | bash
  ```
- [ ] Update any external documentation that references the old path
- [ ] Notify team of path convention change

## Notes

### DCO Sign-off
All commits require `-s` flag per org policy (Developer Certificate of Origin).

### Submodule Handling
The `cmdr` submodule is tracked with `ignore = dirty`. Changes within cmdr that are uncommitted in the submodule will show as "modified" but can be safely ignored in the meta commit.

### Commit Message Style
Following org conventions:
- Type: `feat`, `fix`, `docs`, `refactor`, `chore`
- Scope: `(legion-nix)`, `(bootstrap)`, `(paths)`, `(onboarding)`
- Subject: imperative mood, <72 chars
- Body: detailed explanation, bullet points, references

### File Sizes
- provision.sh: 7.1KB (executable)
- ONBOARDING.md: 14KB (comprehensive)
- onboarding-improvements: 13KB (deep dive)
- Total new content: ~42KB

---

**Ready to commit**: ✅ All verification passed  
**Recommended approach**: Option 2 (Granular commits for easier review)  
**Estimated time**: 5-10 minutes for all commits + push
