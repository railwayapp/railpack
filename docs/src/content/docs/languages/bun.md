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

## Runtime Variables

These variables are available at runtime and build time:

```sh
NODE_ENV=production
NPM_CONFIG_PRODUCTION=false
NPM_CONFIG_UPDATE_NOTIFIER=false
NPM_CONFIG_FUND=false
NPM_CONFIG_FETCH_RETRIES=5
BUN_INSTALL_GLOBAL_STORE=0
CI=true
```

## Dependency Installation

Dependencies are installed using `bun install --frozen-lockfile`.

### Global Store (`BUN_INSTALL_GLOBAL_STORE`)

When using Bun's [isolated linker](https://bun.com/docs/pm/isolated-installs), Bun supports an optional [global virtual store](https://bun.com/docs/pm/global-store) (`globalStore = true` in `bunfig.toml`), which materializes packages into `/root/.bun/install/cache/links/` outside the project. In container builds, this cache directory is an ephemeral cache mount and does not persist across container layers.

To ensure `node_modules` remains self-contained and persistent across build steps and runtime images, `BUN_INSTALL_GLOBAL_STORE=0` is automatically set whenever Bun is the detected package manager. This keeps the isolated store project-local inside `node_modules/.bun/` so that all dependency symlinks remain valid across image layers. Refer to the Bun documentation for official [linker options and recommendations](https://bun.com/docs/pm/isolated-installs#when-to-use-isolated-installs).

## Node.js Compatibility

Some Bun projects also require Node.js for Corepack, scripts, frameworks, or
native module compilation. Railpack installs Node.js automatically when it is
needed.

Bun applications otherwise use the same framework, monorepo, and SPA detection
described in the [Node.js documentation](/languages/node).

