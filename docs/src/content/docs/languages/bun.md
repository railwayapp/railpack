---
title: Bun
description: Building Bun applications with Railpack
---

Railpack supports Bun as a JavaScript runtime and package manager.

## Detection

Railpack selects Bun through JavaScript package manager detection. Bun is
selected when the `packageManager` field declares it, or when no
higher-priority package manager is configured and an `engines.bun` field,
`bun.lock`, or `bun.lockb` file exists.

Bun is also installed when a package script or configured start command invokes
it.

## Versions

The Bun version can be configured using:

- The `RAILPACK_BUN_VERSION` environment variable
- The `packageManager` field in `package.json`, such as `bun@1.2.0`
- The `engines.bun` field in `package.json`
- A `.bun-version`, `mise.toml`, or `.tool-versions` file

The version defaults to `latest` when none is configured.

## Node.js Compatibility

Some Bun projects also require Node.js for Corepack, scripts, frameworks, or
native module compilation. Railpack installs Node.js automatically when it is
needed.

Bun applications otherwise use the same framework, monorepo, and SPA detection
described in the [Node.js documentation](/languages/node).
