# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

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

- `mise run build` - Build the CLI binary
- `mise run check` - Run linting, formatting, and static analysis
- `mise run test` - Run unit tests

# Code style

- Follow Go conventions and existing patterns in the codebase
- Use appropriate error handling with proper error wrapping
- Do not write comments that are obvious from the code itself; focus on
  explaining why something is done, not what it does
- Seriously, do not write comments that are obvious from the code itself.

# Workflow

- Be sure to run `mise run check` when you're done making code changes
- Prefer running focused tests over the full test suite for performance
