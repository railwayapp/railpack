---
title: Node.js
description: Building Node.js applications with Railpack
---

Railpack builds and deploys Node.js applications with support for various
package managers and frameworks.

## Detection

Your project will be detected as a Node.js application if a `package.json` file
exists in the root directory.

## Versions

The Node.js version is determined in the following order of priority:

1. Set via the `RAILPACK_NODE_VERSION` environment variable
2. Read from the `engines.node` field in `package.json`
3. Read from the `.nvmrc` file
4. Read from the `.node-version` file
5. Read from `mise.toml` or `.tool-versions` files
6. Defaults to `22`

This version resolution logic is applied consistently across all scenarios where
Node is needed, including when Bun is the primary package manager but Node is
required for native module compilation.

We officially support actively maintained [Node.js LTS
versions](https://nodejs.org/en/about/previous-releases). Older versions of Node.js will likely still
work but are not officially supported.

### Bun

The Bun version is determined in the following order:

- Set via the `RAILPACK_BUN_VERSION` environment variable
- Read from the `.bun-version` file
- Read from the `engines.bun` field in `package.json`
- Read from `mise.toml` or `.tool-versions` files
- Defaults to `latest`

If Bun is used as the package manager, Node.js will still be installed in the
following cases:

- If you define a `packageManager` field in your `package.json` (for Corepack
  support)
- If any script in your `package.json` contains `node`
- If you're using Astro or Vite
- During installation for native module compilation (node-gyp)

When Node.js isn't required in the final image but is needed during installation
(for native modules), Node.js will be installed via Mise and will respect
[version specifications](#versions).

## Runtime Variables

These variables are available at runtime:

```sh
NODE_ENV=production
NPM_CONFIG_PRODUCTION=false
NPM_CONFIG_UPDATE_NOTIFIER=false
NPM_CONFIG_FUND=false
YARN_PRODUCTION=false
CI=true
```

## Configuration

Railpack builds your Node.js application based on your project structure. The
build process:

- Installs dependencies using your preferred package manager (npm, yarn, pnpm,
  or bun)
- Executes the build script if defined in `package.json`
- Sets up the start command based on your project configuration

Railpack determines the start command in the following order:

1. The `start` script in `package.json`
2. The `main` field in `package.json`
3. An `index.js` or `index.ts` file in the root directory

### Config Variables

| Variable                         | Description                             | Example                                 |
| -------------------------------- | --------------------------------------- | --------------------------------------- |
| `RAILPACK_NODE_VERSION`          | Override the Node.js version            | `22`                                    |
| `RAILPACK_BUN_VERSION`           | Override the Bun version                | `1.2`                                   |
| `RAILPACK_NO_SPA`                | Disable SPA mode                        | `true`                                  |
| `RAILPACK_SPA_OUTPUT_DIR`        | Directory containing built static files | `dist`                                  |
| `RAILPACK_PRUNE_DEPS`            | Remove development dependencies         | `true`                                  |
| `RAILPACK_NODE_PRUNE_CMD`        | Custom command to prune dependencies    | `npm prune --omit=dev --ignore-scripts` |
| `RAILPACK_NODE_INSTALL_PATTERNS` | Custom patterns to install dependencies | `prisma`                                |
| `RAILPACK_ANGULAR_PROJECT`       | Name of the Angular project to build    | `my-app`                                |

### Package Managers

Railpack detects your package manager in the following order:

1. **packageManager field**: Reads the `packageManager` field from
   `package.json` (uses Corepack to install the specified version)
2. **Lock files**: Falls back to detecting based on lock files:
   - `pnpm-lock.yaml` for pnpm
   - `bun.lockb` or `bun.lock` for Bun
   - `.yarnrc.yml` or `.yarnrc.yaml` for Yarn Berry (2+)
   - `yarn.lock` for Yarn 1
3. **engines field**: As a fallback, checks the `engines` field in
   `package.json` for package manager versions:
   - `engines.pnpm` for pnpm version
   - `engines.bun` for Bun version
   - `engines.yarn` for Yarn version
   - Defaults to npm if no package manager is detected

When the `packageManager` field is present, Railpack will use Corepack to
install the specified package manager version. When a package manager is
detected via the `engines` field, the specified version constraint will be
used.

### Monorepo Support

Railpack automatically supports monorepo (workspaces) configurations with all major
package managers. No special configuration is required. 

**Supported Approaches:**

- **npm, bun, yarn**: Uses the `workspaces` field in `package.json`
- **pnpm**: Uses `pnpm-workspace.yaml` configuration
  ([example][pnpm-workspaces-example])

When building a monorepo, Railpack will:

- Detect workspace configurations automatically
- Install all workspace dependencies correctly
- Respect workspace dependency links between packages
- Cache workspace node_modules appropriately

If your monorepo requires building a specific workspace package, ensure
your build and start scripts are defined in the root `package.json` or use
a [config file](/architecture/user-config) to specify custom commands.

### Install

Railpack will only include the necessary files to install dependencies in order
to improve cache hit rates. This includes the `package.json` and relevant lock
files, but there are also a few additional framework specific files that are
included if they exist in your app. This behavior is disabled if a `preinstall`
or `postinstall` script is detected in the `package.json` file.

You can include additional files or directories to include by setting the
`RAILPACK_NODE_INSTALL_PATTERNS` environment variable. This should be a space
separated list of patterns to include. Patterns will automatically be prefixed
with `**/` to match nested files and directories.

## Static Sites

Railpack can serve a statically built Node project with zero config. You can
disable this behavior by either:

- Setting the `RAILPACK_NO_SPA=1` environment variable
- Setting a custom start command

These frameworks are supported:

- **Vite**: Detected if `vite.config.js` or `vite.config.ts` exists, or if the
  build script contains `vite build`
- **Astro**: Detected if `astro.config.js` exists and the output is not type
  `"server"`
- **CRA**: Detected if `react-scripts` is in dependencies and build script
  contains `react-scripts build`
- **Angular**: Detected if `angular.json` exists
- **React Router**: Detected if `react-router.config.js` or
  `react-router.config.ts` exists, or if the build script contains
  `react-router build`. To enable SPA mode, set `ssr: false` in your React
  Router config.

For all frameworks, Railpack will try to detect the output directory and will
default to `dist` (or `build/client/` for React Router). Set the
`RAILPACK_SPA_OUTPUT_DIR` environment variable to specify a custom output
directory.

Static sites are served using the [Caddy](https://caddyserver.com/) web server
and a [default
Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/node/Caddyfile.template).
You can overwrite this file with your own Caddyfile at the root of your project.

## Framework Support

Railpack detects and configures caches and commands for popular frameworks.
Including:

- Next.js: Caches `.next/cache` for each Next.js app in the workspace
- Remix: Caches `.cache`
- Vite: Caches `node_modules/.vite`
- Tanstack Start: Caches `node_modules/.vite`
- Astro: Caches `node_modules/.astro`
- React Router: Caches `.react-router`
- Nuxt:
  - Start command defaults to `node .output/server/index.mjs`
  - Caches `node_modules/.cache`

As well as a default cache for node modules:

- Node modules: Caches `node_modules/.cache` (with the cache key `node-modules`)

## Cache & Removing `node_modules`

When you add custom build commands that remove `node_modules` (such as
`npm ci`), Railpack automatically detects this and
removes the `node_modules/.cache` directory from the cache configuration for
those steps. This prevents `EBUSY: resource busy or locked` [errors](https://github.com/railwayapp/railpack/issues/255)
that would otherwise occur when trying to remove a cached directory.

This automatic handling applies to build steps that contain commands like:

- `npm ci`
- `rm -rf node_modules`
- `rimraf node_modules`

The install step always retains its cache configuration regardless of the
commands used.

### System Dependencies

Railpack automatically installs system dependencies for certain packages:

- **Puppeteer**: When detected in workspace dependencies, Railpack installs
  all necessary system packages for running headless Chrome, including
  `xvfb`, `chromium` dependencies, and font libraries
