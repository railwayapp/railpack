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

# Code style

- Follow Go conventions and existing patterns in the codebase
- Use appropriate error handling with proper error wrapping
- Do not write comments that are obvious from the code itself; focus on
  explaining why something is done, not what it does
- Seriously, do not write comments that are obvious from the code itself.
- Do not write one-line functions
- Always use the App abstraction for file system operations.

# Workflow

- Do not use `go` directly. Instead, inspect @mise.toml and use `mise run <task>` to run various dev lifecycle commands.
- Be sure to run `mise run check` when you're done making code changes
- Run unit tests and a couple of relevant integration tests to verify your changes
  - Don't run tests manually using `go test` unless instructed to do so
  - If tests are failing that are unrelated to your changes, let me know and stop working.
- Use the `cli` mise task to test your changes on a specific example project, i.e. `mise run cli -- --verbose build --show-plan examples/node-vite-react-router-spa/`
- Do not run any write operations with `git`
- Do not use `bin/railpack` instead use `mise run cli` (which is the development build of `railpack`)

# File Conventions

- Markdown files in @docs/src/content/docs/ should be limited to 80 columns
