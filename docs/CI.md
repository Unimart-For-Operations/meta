# CI/CD Pipeline - Gitea Actions

## Overview

The meta repository uses Gitea Actions for continuous integration. The pipeline validates code quality, documentation consistency, and cross-repo contracts on every push and pull request.

## Pipeline Location

`.gitea/workflows/ci.yml`

## Trigger Events

- **Push** to `main` branch
- **Pull requests** targeting `main` branch

## Jobs Overview

### 1. ShellCheck (Scripts)

**Purpose**: Validate shell script quality and catch common bugs.

**Checks**:
- `scripts/provision.sh`
- `scripts/setup.sh`
- `cmdr/scripts/bootstrap.sh`
- All `.sh` files in the repository

**Tools**: [ShellCheck](https://www.shellcheck.net/)

**Exit Criteria**:
- ✅ No ShellCheck warnings or errors
- ✅ All scripts use proper error handling (`set -euo pipefail`)
- ✅ Proper quoting of variables

### 2. Go (fmt, vet, build, test)

**Purpose**: Validate Go code quality and ensure unimart builds successfully.

**Checks**:
- `go fmt` on meta module
- `go fmt` on idpbuilder nested module
- `go vet` on both modules
- Build unimart binary with version injection
- Run unit tests on both modules

**Tools**: Go 1.23 toolchain

**Exit Criteria**:
- ✅ All Go code is properly formatted
- ✅ No `go vet` warnings
- ✅ unimart binary builds successfully
- ✅ All unit tests pass

### 3. Nix (flake check)

**Purpose**: Validate Nix flake configurations and host definitions.

**Checks**:
- `nix flake check --all-systems` (evaluates all hosts)
- `nix fmt` check (ensures Nix code is formatted)

**Tools**: Nix with flakes enabled

**Exit Criteria**:
- ✅ All flake outputs evaluate successfully
- ✅ All host configurations are syntactically valid
- ✅ Nix code is properly formatted with nixfmt

### 4. Documentation (markdown, paths)

**Purpose**: Validate documentation quality and path consistency.

**Checks**:
- Markdown linting (markdownlint-cli)
- Path consistency check (no stray `repos/github/idpbuilder`)
- provision.sh existence and executable bit

**Tools**: markdownlint-cli, grep

**Exit Criteria**:
- ⚠️  Markdown lint issues are reported but non-fatal
- ✅ No references to old path convention
- ✅ provision.sh exists and is executable

**Path Check Details**:
```bash
# Searches for old path pattern in:
grep -r "repos/github/idpbuilder" --include="*.md" --include="*.sh" --include="*.nix"

# Expected: 0 matches (all paths updated to repos/Unimart-For-Operations/meta)
```

### 5. Host Configs (NixOS)

**Purpose**: Validate all NixOS host configurations can be evaluated.

**Checks**:
- Verifies required files exist (meta.nix, default.nix, system.nix, hardware-configuration.nix)
- Evaluates each host configuration (syntax check, no build)

**Tools**: Nix with flakes enabled

**Exit Criteria**:
- ✅ All NixOS hosts have required files
- ✅ All NixOS configurations evaluate successfully
- ✅ No syntax errors in host configs

**Hosts Validated**:
- `legion-nix`
- `strix-nix`
- `nixos`
- (Future NixOS hosts automatically included)

### 6. Cross-Repo Contracts

**Purpose**: Validate contracts between meta and submodules.

**Checks**:
- Runs `unimart stockroom check`
- Validates submodule structure
- Checks for drift

**Tools**: unimart (built from source)

**Exit Criteria**:
- ✅ All cross-repo contracts satisfied
- ✅ No unexpected drift detected

### 7. Provision Script (smoke test)

**Purpose**: Validate provision.sh syntax and basic operation.

**Checks**:
- Bash syntax check (`bash -n`)
- (Future: dry-run mode test)

**Tools**: bash

**Exit Criteria**:
- ✅ provision.sh has valid bash syntax
- ✅ No syntax errors

### 8. CI Summary

**Purpose**: Final gate - requires all jobs to pass.

**Exit Criteria**:
- ✅ All 7 jobs completed successfully

## Running CI Locally

### Prerequisites

Install required tools:

```bash
# macOS
brew install shellcheck markdownlint-cli go

# Arch Linux
sudo pacman -S shellcheck nodejs npm go
sudo npm install -g markdownlint-cli

# NixOS/Nix
nix-shell -p shellcheck nodejs go
npm install -g markdownlint-cli
```

### Run All Checks

```bash
# From meta repository root
make ci
```

This delegates to `unimart stockroom check` which runs the full validation suite.

### Run Individual Checks

**ShellCheck**:
```bash
find scripts/ -type f -name "*.sh" -exec shellcheck {} \;
find cmdr/scripts/ -type f -name "*.sh" -exec shellcheck {} \;
```

**Go checks**:
```bash
go fmt ./...
go -C idpbuilder fmt ./...
go vet ./...
go -C idpbuilder vet ./...
go build -o unimart .
go test ./...
go -C idpbuilder test ./...
```

**Nix checks**:
```bash
cd cmdr
nix flake check --all-systems
nix fmt -- --check .
```

**Markdown lint**:
```bash
markdownlint '**/*.md' --ignore node_modules --ignore .git --config .markdownlint.json
```

**Path consistency**:
```bash
grep -r "repos/github/idpbuilder" \
  --include="*.md" \
  --include="*.sh" \
  --include="*.nix" \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  .
# Expected: no output (0 matches)
```

**Host configs**:
```bash
cd cmdr
for host in home/02-hosts/nixos/*/; do
  host_name=$(basename "$host")
  [ "$host_name" = "_template" ] && continue
  echo "Checking $host_name..."
  nix eval ".#nixosConfigurations.${host_name}.config.system.build.toplevel" > /dev/null
done
```

## CI Performance

Approximate job durations (on GitHub Actions runners):

| Job | Duration | Cacheable |
|-----|----------|-----------|
| ShellCheck | ~30s | No |
| Go checks | ~2-3min | Yes (Go cache) |
| Nix checks | ~5-10min | Yes (Nix cache) |
| Documentation | ~1min | No |
| Host configs | ~3-5min | Yes (Nix cache) |
| Cross-repo | ~1min | Yes (Go cache) |
| Provision test | ~30s | No |
| **Total** | **~15-20min** | |

With caching, subsequent runs: **~5-8min**

## Failure Scenarios

### ShellCheck Failures

**Common Issues**:
- Unquoted variables
- Missing error handling
- Bash-specific syntax in `/bin/sh` scripts

**Fix**:
```bash
shellcheck scripts/provision.sh
# Apply suggested fixes
```

### Go Build Failures

**Common Issues**:
- Unformatted code
- Unused imports
- Type errors
- Test failures

**Fix**:
```bash
go fmt ./...
go vet ./...
go build -o unimart .
go test ./...
```

### Nix Evaluation Failures

**Common Issues**:
- Syntax errors in .nix files
- Missing imports
- Invalid attribute names
- Host configuration errors

**Fix**:
```bash
cd cmdr
nix flake check --all-systems
# Review error output
# Fix issues in indicated files
nix fmt .
```

### Path Consistency Failures

**Common Issues**:
- Documentation still references old path (`repos/github/idpbuilder`)

**Fix**:
```bash
# Find all references
grep -rn "repos/github/idpbuilder" . --include="*.md"

# Update to new path
sed -i 's|repos/github/idpbuilder|repos/Unimart-For-Operations/meta|g' <file>
```

### Host Config Failures

**Common Issues**:
- Missing required files (meta.nix, system.nix, etc.)
- Syntax errors in Nix configuration
- Invalid hardware-configuration.nix

**Fix**:
```bash
# Validate host manually
cd cmdr
nix eval ".#nixosConfigurations.legion-nix.config.system.build.toplevel"

# Check for syntax errors in indicated files
```

## Bypassing CI (Not Recommended)

In rare cases where CI must be bypassed (e.g., emergency hotfix):

```bash
# Commit with [skip ci] in message
git commit -m "hotfix: emergency fix [skip ci]"
```

**⚠️  Warning**: Bypassing CI should be extremely rare and requires post-merge validation.

## CI Configuration Updates

When updating CI workflows:

1. **Local validation first**:
   ```bash
   # Validate YAML syntax
   yamllint .gitea/workflows/ci.yml
   
   # Test changes don't break existing checks
   make ci
   ```

2. **Test on feature branch**:
   - Create PR with CI changes
   - Verify all jobs run successfully
   - Review job logs for issues

3. **Document changes**:
   - Update this file (docs/CI.md)
   - Add entry to CHANGELOG if significant

## Monitoring & Debugging

### View CI Logs (Gitea)

1. Navigate to: `https://gitea.cnoe.localtest.me/<org>/meta/actions`
2. Select workflow run
3. Click on job to view logs

### Common Debug Steps

**Job stuck or timeout**:
- Check for infinite loops in scripts
- Review Nix evaluation for circular dependencies
- Check for network issues (flake fetches)

**Intermittent failures**:
- Check for race conditions
- Review caching behavior
- Verify external dependencies (GitHub, Nix channels)

**Cache issues**:
- Clear runner cache
- Rebuild from scratch
- Check cache key configuration

## Future Enhancements

### Short-Term
- [ ] Add dry-run mode to provision.sh for testing
- [ ] Add container-based provision.sh smoke test
- [ ] Cache Nix store between runs
- [ ] Add Go test coverage reporting

### Medium-Term
- [ ] Add security scanning (gosec, gitleaks)
- [ ] Add dependency vulnerability scanning
- [ ] Parallel job execution optimization
- [ ] Add performance regression testing

### Long-Term
- [ ] Multi-platform testing (macOS, Arch, NixOS, Ubuntu)
- [ ] Integration tests for full bootstrap flow
- [ ] Automated release tagging
- [ ] Deploy preview environments for PRs

## References

- **Gitea Actions**: https://docs.gitea.com/usage/actions/overview
- **ShellCheck**: https://www.shellcheck.net/
- **markdownlint**: https://github.com/DavidAnson/markdownlint
- **Nix flake check**: https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-check

## Support

For CI issues:
1. Check job logs in Gitea Actions
2. Reproduce locally with `make ci`
3. Review this documentation
4. Check for similar issues in org repos
5. Create issue with `[CI]` prefix if needed

---

**Last Updated**: 2026-08-31  
**Maintainer**: cmdr  
**Status**: Active
