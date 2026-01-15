# What is Railpack

Zero-config application builder that automatically analyzes your code and turns
it into a container image. It's built on BuildKit
with support for Node, Python, Go, PHP, and more.

# Architecture

- **Core**: Analyzes apps and generates JSON build plans using language
  providers
- **BuildKit**: Converts build plans to BuildKit LLB (Low-Level Builder) format
  for efficient image construction
- **CLI**: Main entry point that coordinates core analysis and BuildKit
  execution
- **Providers**: Language-specific modules that detect project types (e.g. Node
  detects package.json) and generate appropriate build steps
- **Runtime**: The built images are based on @images/debian/runtime/Dockerfile

# Code style

- Follow Go conventions and existing patterns in the codebase
- Use appropriate error handling with proper error wrapping
- Do not write comments that are obvious from the code itself; focus on
  explaining why something is done, not what it does
- Seriously, do not write comments that are obvious from the code itself.
- Do not write one-line functions
- Always use the App abstraction for file system operations.
- When configuring Mise settings, prefer environment variables (e.g.
  `MISE_PIPX_UVX`) over TOML settings to maintain consistency with existing
  patterns in the codebase.

# Mise Setup

If mise is not available in your environment, install it using the instructions at https://mise.jdx.dev/installing-mise.html.

# Workflow

- Take a careful look at @mise.toml to understand what commands should be run at different points in the project lifecycle
- Do not worry about docker cache, etc. Never run `docker system prune` or any other similar commands.
- Do not run `go` directly. Instead, inspect @mise.toml and use `mise run <task>` to run various dev lifecycle commands. For instance, you should not run `go vet`, `go fmt`, `go test`, etc directly.
- After making code changes, first run `mise run check`
- Then, run unit tests and a couple of relevant integration tests to verify your changes
  - Don't run tests manually using `go test` unless instructed to do so
  - If tests are failing that are unrelated to your changes, let me know and stop working.
- Use the `cli` mise task to test your changes on a specific example project, i.e. `mise run cli -- --verbose build --show-plan examples/node-vite-react-router-spa/`
- Do not run any write operations with `git`
- Do not use `bin/railpack` instead use `mise run cli` (which is the development build of `railpack`)
  - Therefore do not run `mise build`, we don't need a `railpack` binary for local testing
- Use a local `tmp/` directory for temporary files instead of the global `/tmp/`
  directory. Don't worry about cleaning up tmp directories.
- You can clone the mise repo (https://github.com/jdx/mise) into `tmp/` to
  inspect it when needed.

# Tests

There are normal unit tests, snapshot tests, and integration tests. The integration tests are most unique to this project:

* They represent example projects that would be built using the `railpack` CLI
* On CI, they are built and run to make sure `railpack` properly builds *and* runs the project
* `test.json` and `docker-compose.yml` are used to help determine what assertions should be made and what services should be run for the test

## Integration Tests

* Run `mise run test-integration-cwd` from within an `examples/` directory to run the integration test for that example.
* `test.json` files are JSONC format and support comments. Use comments to explain temporary workarounds or special test conditions.
* In `test.json` we should avoid using `justBuild` for all but the most simple projects. `justBuild` does not test `expectedOutput` or any other assertions.
* If the project has a server component, we should use a `httpCheck` test. Read the @docs/src/content/docs/guides/developing-locally.md guide, specifically the `### HTTP Checks` section for more information.
* `httpCheck` assertions assume that `$PORT` is respected.
* You can use `"env": { "SECRET": "123"}` to add a required environment variable to a test case.

# File Conventions

- Markdown files in @docs/src/content/docs/ should be limited to 80 columns
- Do not fix indentation or formatting manually. This is corrected automatically using `mise run check`
- NEVER commit language-specific cache or build artifacts (e.g. Python `__pycache__/`, `*.pyc`, `*.pyo` files). If you encounter build artifacts that are not excluded by .gitignore, add appropriate patterns to .gitignore to exclude them