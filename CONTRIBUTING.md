# Contributing to Railpack

## Pull Requests

We welcome pull requests that push the project forward in meaningful ways.
Please ensure your PRs:

- Address a specific problem or add a well-defined feature
- Include tests for new functionality
- Follow the existing code style

Note: We prefer focused, well-thought-out contributions over "drive-by" PRs that
make superficial changes.

DO NOT submit completely AI-generated PRs. If we suspect PRs are completely generated,
we will close them with a note indicating this (feel free to reopen if we got it wrong!).

## Development Workflow

[Checkout this guide for more information.](https://railpack.com/guides/developing-locally/)

## Testing

### Core Tests

- All example plans are snapshot tested in `core_test.go`
- Tests with a `test.json` file will be built and run automatically
- The test output must contain the `expectedOutput` specified in the test file

### Snapshot Tests

Railpack uses [go-snaps](https://github.com/gkampitakis/go-snaps) for snapshot
testing. This helps prevent regressions to generated build plans.

If you see a test failure because of a snapshot change, please confirm that the
change is intentional, and then update the snapshot by running:

```bash
mise run test-update-snapshots
```

### Integration Tests

Example directories with a `test.json` file will be automatically built and run
in CI. You can run them locally with:

```bash
mise run test-integration
```

The `test.json` file contains an array of build configurations and expected
outputs. See [this
file](https://github.com/railwayapp/railpack/blob/main/integration_tests/run_test.go#L26)
for the schema.

## Useful Commands

- `mise check` - Run linting and type checking
- `mise test` - Run unit tests
- `mise test-integration` - Run integration tests
- `mise run test-update-snapshots` - Update snapshot tests
