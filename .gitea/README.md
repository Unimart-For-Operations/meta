# .gitea Directory

This directory contains Gitea-specific configuration for the meta repository.

## Structure

```
.gitea/
└── workflows/
    └── ci.yml        # Main CI pipeline (runs on push/PR)
```

## Workflows

### ci.yml - Continuous Integration

**Triggers**: Push to main, Pull requests to main

**Jobs**:
1. ShellCheck - Validate shell scripts
2. Go checks - fmt, vet, build, test (meta + idpbuilder)
3. Nix checks - flake check, fmt validation
4. Documentation - markdown lint, path consistency
5. Host configs - validate NixOS configurations
6. Cross-repo - stockroom contract validation
7. Provision test - provision.sh smoke test
8. CI summary - final gate (all jobs must pass)

**Documentation**: See [/docs/CI.md](../docs/CI.md) for complete details.

## Adding New Workflows

1. Create YAML file in `.gitea/workflows/`
2. Follow Gitea Actions syntax: https://docs.gitea.com/usage/actions/overview
3. Test locally before pushing
4. Document in `/docs/CI.md`

## Local Testing

Run all CI checks locally:

```bash
make ci
```

Or run individual checks (see `/docs/CI.md` for commands).

## CI Status Badge

(To be added once Gitea is configured)

```markdown
[![CI Status](https://gitea.cnoe.localtest.me/<org>/meta/actions/workflows/ci.yml/badge.svg)](https://gitea.cnoe.localtest.me/<org>/meta/actions)
```

## References

- [Gitea Actions Documentation](https://docs.gitea.com/usage/actions/overview)
- [CI Documentation](/docs/CI.md)
- [Meta Repository AGENTS.md](/AGENTS.md)
