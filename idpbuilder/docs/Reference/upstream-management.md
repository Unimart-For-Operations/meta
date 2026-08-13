# Upstream Management

This repo is a private, independent copy of [cnoe-io/idpbuilder](https://github.com/cnoe-io/idpbuilder). The GitHub fork relationship was intentionally broken to allow private hosting.

## Setup

The `upstream` remote should point to the original project:

```bash
git remote add upstream git@github.com:cnoe-io/idpbuilder.git
git fetch upstream --tags
```

Verify:
```bash
git remote -v
# origin    git@github.com:Unimart-For-Operations/idpbuilder.git (fetch)
# origin    git@github.com:Unimart-For-Operations/idpbuilder.git (push)
# upstream  git@github.com:cnoe-io/idpbuilder.git (fetch)
# upstream  git@github.com:cnoe-io/idpbuilder.git (push)
```

## Tracking Upstream

```bash
# Fetch latest upstream changes and tags
make fetch-upstream

# Show how far ahead/behind we are
make upstream-status

# List upstream commits not yet in our main branch
make log-upstream

# Show upstream commits with full diffs
make log-upstream-detail

# Show a stat diff between our main and upstream
make diff-upstream
```

## Cherry-Picking

When upstream has changes we want:

```bash
# 1. Fetch latest
make fetch-upstream

# 2. Review what's new
make log-upstream

# 3. Cherry-pick a single commit
make cherry-pick COMMIT=abc123

# 4. Cherry-pick a range of commits
make cherry-pick-range FROM=abc123 TO=def456
```

### Resolving Conflicts

Cherry-picks may conflict. When they do:

1. Git will pause the cherry-pick and show conflicting files
2. Resolve conflicts in your editor
3. `git add` the resolved files
4. `git cherry-pick --continue`
5. If you want to abort: `git cherry-pick --abort`

## Go Module Path

The Go module path is `github.com/cnoe-io/idpbuilder` — this intentionally matches upstream to avoid rewriting every import in the codebase. References to `cnoe-io` in Go source files, `go.mod`, and Makefile LD_FLAGS are expected and correct.

## What Changed from Upstream

To see all commits unique to this fork:

```bash
git log --oneline upstream/main..HEAD
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `fetch-upstream` | Fetch latest changes and tags from cnoe-io/idpbuilder |
| `log-upstream` | Show upstream commits not in our main branch |
| `log-upstream-detail` | Show upstream commits with full diffs |
| `diff-upstream` | Show stat diff between our main and upstream/main |
| `cherry-pick COMMIT=<sha>` | Cherry-pick a specific upstream commit |
| `cherry-pick-range FROM=<sha> TO=<sha>` | Cherry-pick a range of upstream commits |
| `upstream-status` | Show ahead/behind counts and last fetch time |

## Why the Fork Was Broken

GitHub enforces that forks of public repos must remain public. To make this repo private:

1. The original fork (`idpbuilder/idpbuilder-old`) was renamed
2. A new private repo (`idpbuilder/idpbuilder`) was created
3. All branches and tags were pushed from the local clone
4. The old repo was deleted
5. An `upstream` remote was added pointing to `cnoe-io/idpbuilder`

All 259 commits and 16 tags from the original project history are preserved.
