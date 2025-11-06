---
title: Deno
description: Building Deno applications with Railpack
---

Railpack builds and deploys Deno applications with zero configuration.

## Detection

Your project will be detected as a Deno application if a `deno.json` or
`deno.jsonc` file exists in the root directory.

## Versions

The Deno version is determined in the following order:

- Set via the `RAILPACK_DENO_VERSION` environment variable
- Read from mise-compatible version files (`.deno-version`,
  `.tool-versions`, `mise.toml`)
- Defaults to `2`

## Configuration

Railpack builds your Deno application based on your project structure. The build
process:

- Installs Deno
- Caches dependencies using `deno cache`
- Sets up the start command based on your project configuration

The start command is determined by looking for:

1. A `main.ts`, `main.js`, `main.mjs`, or `main.mts` file in the project root
2. If no main file is found, it will use the first `.ts`, `.js`, `.mjs`, or
   `.mts` file found in your project

The selected file will be run with `deno run --allow-all`.

### Config Variables

| Variable                | Description               | Example |
| ----------------------- | ------------------------- | ------- |
| `RAILPACK_DENO_VERSION` | Override the Deno version | `1.41`  |
