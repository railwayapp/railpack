# Agents

## Cursor Cloud specific instructions

Railpack is a Go CLI tool that builds container images from source code using BuildKit. There is no web UI or running server — development is centered on the CLI and its test suite.

### Prerequisites

Docker and BuildKit must be running before any build or integration test:

```bash
# Start Docker daemon (if not already running)
dockerd &>/tmp/dockerd.log &
# Start BuildKit container
docker run --rm --privileged -d --name buildkit -e BUILDKIT_DEBUG=1 moby/buildkit:latest || true
```

### Key commands

All dev tasks use `mise run <task>` — see `mise.toml` for the full list. Do **not** run `go` commands directly.

| Task | Command |
|---|---|
| Lint/format/vet | `mise run check` |
| Unit tests | `mise run test` |
| Integration test (single example) | `cd examples/<name> && mise run test-integration-cwd` |
| Build an example (show plan) | `mise run cli -- --verbose build --show-plan examples/<name>/` |
| Update snapshots | `mise run test-update-snapshots` |

### Gotchas

- The `BUILDKIT_HOST` env var is set automatically by `mise.toml` (`docker-container://buildkit`). If BuildKit isn't running, builds will hang or fail without a clear error.
- Integration tests require Docker, BuildKit, and network access to pull base images from `ghcr.io/railwayapp/railpack-*`.
- Some snapshot tests track upstream package versions (e.g. PHP FrankenPHP image tags) and may fail when upstream publishes a new patch. Update snapshots with `mise run test-update-snapshots`.
- The docs site (`docs/`) uses Bun + Astro and is independent from the Go codebase. Run `mise run docs-dev` to start the docs dev server.
