# repositories

Designated local source directory for repos published to in-cluster Gitea.

When `unimart open` runs, it publishes git repos found in this directory by default.

## Contents

Each entry is a symlink to the corresponding submodule directory at the org root:

| Symlink | Target | Gitea repo |
|---------|--------|------------|
| `cmdr` | `../cmdr` | `idpbuilder/cmdr` |
| `idpbuilder` | `../idpbuilder` | `idpbuilder/idpbuilder` |
| `meta` | `../` (meta itself) | `idpbuilder/meta` |

Symlinks are used to avoid duplicate clones — the submodule directories at the org root
are the canonical local copies. `unimart` resolves symlinks when scanning for `.git` entries.

## Usage

```bash
# Dry-run to see what would be published (default)
unimart freezer repos publish-to-gitea --gitea-url https://gitea.cnoe.localtest.me:8443

# Full publish (runs automatically during unimart open)
unimart freezer repos publish-to-gitea --gitea-url https://gitea.cnoe.localtest.me:8443 --dry-run=false
```

## Adding a repo

1. Add a symlink: `ln -s <relative-path-to-repo> repositories/<name>`
2. The repo will be published to Gitea on the next `unimart open` or manual publish run.

## Notes

- Entries must resolve to directories containing a `.git` entry (file or directory).
- Symlinks to directories are supported — `unimart` uses `os.Stat` (follows symlinks).
- This directory can contain regular clones, worktrees, or submodule checkouts.
- If `repositories/` does not exist, unimart falls back to legacy org-root scanning.
