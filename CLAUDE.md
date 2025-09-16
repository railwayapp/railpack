# What is Railpack

Zero-config application builder that automatically analyzes your code and turns
it into a container image. It's the successor to Nixpacks, built on BuildKit
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

# Bash commands

- `mise run build` - Build the CLI binary (`./bin/railpack`), use this instead of `go build` (`./bin/railpack`), use this instead of `go build`
- `mise run check` - Run linting, formatting, and static analysis
- `mise run test` - Run unit tests
- `mise run test-integration` - Runs *all* integration tests. Never run this. Instead, only run the test that you are currently working on, e.g. `mise run test-integration -- -run "TestExamplesIntegration/config-file"`

# Code style

- Follow Go conventions and existing patterns in the codebase
- Use appropriate error handling with proper error wrapping
- Do not write comments that are obvious from the code itself; focus on
  explaining why something is done, not what it does
- Seriously, do not write comments that are obvious from the code itself.
- Do not write one-line functions

# Workflow

- Be sure to run `mise run check` when you're done making code changes
- Don't run tests manually unless instructed to do so
- Use the `cli` mise task to test your changes on a specific example project, i.e. `mise run cli -- --verbose build --show-plan examples/node-vite-react-router-spa/`
- Never `git commit`
- Do not use `bin/railpack` instead use `mise run cli` (which is the development build of `railpack`)

# File Conventions

- Markdown files in @docs/src/content/docs/ should be limited to 80 columns
