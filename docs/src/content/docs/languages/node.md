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

The Node.js version is determined in the following order:

- Set via the `RAILPACK_NODE_VERSION` environment variable
- Read from the `engines` field in `package.json`
- Read from the `.nvmrc` file
- Defaults to `22`

### Bun

The Bun version is determined in the following order:

- Set via the `RAILPACK_BUN_VERSION` environment variable
- Defaults to `latest`

If Bun is used as the package manager, Node.js will still be installed in the
following cases:

- If you define a `packageManager` field in your `package.json` (for Corepack
  support)
- If any script in your `package.json` contains `node`
- If you're using Astro

When Node.js isn't required in the final image but is needed during installation
(for native modules), a basic Node.js version will be installed from apt
packages.

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

| Variable                         | Description                             | Example  |
| -------------------------------- | --------------------------------------- | -------- |
| `RAILPACK_NODE_VERSION`          | Override the Node.js version            | `22`     |
| `RAILPACK_BUN_VERSION`           | Override the Bun version                | `1.2`    |
| `RAILPACK_NO_SPA`                | Disable SPA mode                        | `true`   |
| `RAILPACK_SPA_OUTPUT_DIR`        | Directory containing built static files | `dist`   |
| `RAILPACK_PRUNE_DEPS`            | Remove development dependencies         | `true`   |
| `RAILPACK_NODE_INSTALL_PATTERNS` | Custom patterns to install dependencies | `prisma` |
| `RAILPACK_ANGULAR_PROJECT`       | Name of the Angular project to build    | `my-app` |

### Package Managers

Railpack detects your package manager based on lock files:

- `pnpm-lock.yaml` for pnpm
- `bun.lockb` or `bun.lock` for Bun
- `.yarnrc.yml` or `.yarnrc.yaml` for Yarn 2
- `yarn.lock` for Yarn 1
- Defaults to npm if no lock file is found

### Install

Railpack will only include the necessary files to install dependencies in order
to improve cache hit rates. This includes the `package.json` and relevant lock
files, but there are also a few additional framework specific files that are
included if they exist in your app. This behaviour is disabled if a `preinstall`
or `postinstall` script is detected in the `package.json` file.

You can include additional files or directories to include by setting the
`RAILPACK_NODE_INSTALL_PATTERNS` environment variable. This should be a space
separated list of patterns to include. Patterns will automatically be prefixed
with `**/` to match nested files and directories.

## Static Sites

Railpack can serve a statically built Node project with zero config. You can
disable this behaviour by either:

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

For both frameworks, Railpack will try to detect the output directory and will
default to `dist`. Set the `RAILPACK_SPA_OUTPUT_DIR` environment variable to
specify a custom output directory.

Static sites are served using the [Caddy](https://caddyserver.com/) web server
and a [default
Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/node/Caddyfile.template).
You can overwrite this file with your own Caddyfile at the root of your project.

## Framework Support

Railpack detects and configures caches and commands for popular frameworks.
Including:

- Next.js: Caches `.next/cache` for each Next.js app in the workspace
- Remix: Caches `.cache`
- Vite (and Tanstack Start): Caches `.vite/cache`
- Astro: Caches `.astro/cache`
- Nuxt:
  - Start command defaults to `node .output/server/index.mjs`
  - Caches `node_modules/.cache`

As well as a default cache for node modules:

- Node modules: Caches `node_modules/.cache`
