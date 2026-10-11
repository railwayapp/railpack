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

### Node.js

The Node.js version is determined in the following order of priority:

1. Set via the `RAILPACK_NODE_VERSION` environment variable
2. Read from `devEngines.runtime` in `package.json` through
   [Mise's idiomatic file parsing](#mise-idiomatic-file-parsing)
3. Read from the `engines.node` field in `package.json`
4. Read from the `.nvmrc` file
5. Read from the `.node-version` file
6. Read from `mise.toml` or `.tool-versions` files
7. Defaults to `lts`

This version resolution logic is applied consistently across all scenarios where
Node is needed, including when Bun is the primary package manager but Node is
required for native module compilation.

We officially support actively maintained [Node.js LTS
versions](https://nodejs.org/en/about/previous-releases). Older versions of
Node.js will likely still work but are not officially supported.

Node.js GPG verification is disabled by default; see the [GPG verification
recommendation](/config/recommendations#enable-gpg-verification) to enable it in
your project.

### Package Manager Versions

The detected package manager's version is determined in the following order:

1. `package.json`, through Mise's idiomatic file parsing described below.
2. The package manager's `engines` field, such as `engines.pnpm`.
3. The detected lock file, when its format identifies a compatible version.
4. The default version for the detected package manager.

### Mise Idiomatic File Parsing

For Node.js, npm, Yarn, pnpm, and Bun, Mise parses `package.json` as an
[idiomatic version file](https://mise.jdx.dev/configuration.html#idiomatic-version-files).
For Node.js, Mise reads `devEngines.runtime.version` when
`devEngines.runtime.name` is `node`.

For package managers, Mise checks these fields in order:

1. `devEngines.packageManager`: Uses `version` when `name` matches the detected
   package manager. The value may be an object or an array, in which case Mise
   reads the first entry.
2. `packageManager`: Parses `<package-manager>@<version>` and removes an optional
   `+hash` suffix.

Railpack parses additional package manager version sources as described in
[Versions](#package-manager-versions).

## Runtime Variables

These variables are available at runtime:

```sh
NODE_ENV=production
NPM_CONFIG_PRODUCTION=false
NPM_CONFIG_UPDATE_NOTIFIER=false
NPM_CONFIG_FUND=false
NPM_CONFIG_FETCH_RETRIES=5
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
| `RAILPACK_NO_SPA`                | Disable SPA mode                        | `true`                                  |
| `RAILPACK_SPA_OUTPUT_DIR`        | Directory containing built static files | `dist`                                  |
| `RAILPACK_PRUNE_DEPS`            | Remove development dependencies         | `true`                                  |
| `RAILPACK_NODE_NPM_INSTALL`      | Custom npm install command              | `npm ci`                                |
| `RAILPACK_NODE_PRUNE_CMD`        | Custom command to prune dependencies    | `npm prune --omit=dev --ignore-scripts` |
| `RAILPACK_NODE_INSTALL_PATTERNS` | Custom patterns to install dependencies | `prisma`                                |
| `RAILPACK_ANGULAR_PROJECT`       | Name of the Angular project to build    | `my-app`                                |
| `RAILPACK_NX_APP`                | Nx app to build and start (project name, package name, or path) | `web` or `@org/web` |
| `RAILPACK_NODE_PLAYWRIGHT_INSTALL` | Install Playwright browsers | `1` |

### Playwright

When Playwright is a production dependency, Railpack suggests setting
`RAILPACK_NODE_PLAYWRIGHT_INSTALL=1`. Browser installation is opt-in because
it increases image size and is not required by every application that includes
Playwright.

When enabled, Railpack runs Playwright through the detected package manager to
install its browser binaries and adds the required runtime system packages.
Ensure Playwright is included in your production dependencies so its CLI is
available during the build.

### Package Managers

Railpack detects your package manager in the following order:

1. **packageManager field**: Reads the `packageManager` field from
   `package.json`
2. **Mise idiomatic version files**: If Mise resolves exactly one of
   pnpm, Yarn, Bun, or npm from `package.json` (`devEngines.packageManager`
   or `packageManager`), that manager is used. A tool listed only in
   `mise.toml` or `.tool-versions` does not select the package manager.
3. **Lock files**: Falls back to detecting based on lock files:
   - `pnpm-lock.yaml` for pnpm
   - `bun.lockb` or `bun.lock` for Bun
   - `.yarnrc.yml` or `.yarnrc.yaml` for Yarn Berry (2+)
   - `yarn.lock` for Yarn 1
4. **engines field**: As a fallback, checks the `engines` field in
   `package.json` for package manager versions:
   - `engines.pnpm` for pnpm version
   - `engines.bun` for Bun version
   - `engines.yarn` for Yarn version
   - Defaults to npm if no package manager is detected

When the `packageManager` field selects npm or Yarn, Corepack installs its
specified version.

Railpack supports building native modules and automatically configures `node-gyp`.

### Monorepo Support

Railpack automatically supports monorepo (workspaces) configurations with all major
package managers. No special configuration is required.

**Supported Approaches:**

- **npm, bun, yarn**: Uses the `workspaces` field in `package.json`
- **pnpm**: Uses `pnpm-workspace.yaml` configuration
- **Nx**: Detects `nx.json` and builds Next.js apps even when targets are
  inferred (no root `build`/`start` scripts)

See the [examples
folder](https://github.com/railwayapp/railpack/tree/main/examples) in the
repository for workspace examples across different package managers (e.g.,
`node-pnpm-workspaces`, `node-npm-workspaces`, `node-yarn-workspaces`,
`node-bun-workspaces`, `node-nx-next`).

When building a monorepo, Railpack will:

- Detect workspace configurations automatically
- Install all workspace dependencies correctly
- Respect workspace dependency links between packages
- Cache workspace node_modules appropriately

If your monorepo requires building a specific workspace package, ensure
your build and start scripts are defined in the root `package.json`, set
`RAILPACK_NX_APP` for multi-app Nx workspaces, or use a
[configuration file](/config/file) to specify custom commands.

#### Nx

Stock Nx workspaces often rely on [inferred
tasks](https://nx.dev/docs/concepts/inferred-tasks) instead of `package.json`
scripts. When Railpack detects Nx and the root has no `build`/`start` scripts:

- **Build**: `nx build <project>` (uses the package name, e.g.
  `@org/web`)
- **Start** (Next.js): `cd apps/web && next start` so runtime does not depend
  on the `nx` CLI

With a single Next.js app, selection is automatic. With multiple apps, set
`RAILPACK_NX_APP` to the project name, package name, or package path (e.g.
`web`, `@org/web`, or `apps/web`).

#### TanStack Start

TanStack Start is detected via `@tanstack/react-start` and is not treated as a
Vite SPA. If there is no `start` script, Railpack installs `srvx` globally and
starts with `srvx --prod -s ../client dist/server/server.js`. For production
Node deploys, set up Nitro per the
[TanStack hosting docs](https://tanstack.com/start/latest/docs/framework/react/guide/hosting#nitro).
Railpack caches `node_modules/.vite`.

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
- **Next.js**: Detected if `next` is in dependencies and `next.config.js`,
  `next.config.mjs`, or `next.config.ts` sets `output: 'export'` (or
  `output: "export"`). The default `next start` start script does not
  disable SPA mode.
- **CRA**: Detected if `react-scripts` is in dependencies and build script
  contains `react-scripts build`
- **Angular**: Detected if `angular.json` exists
- **React Router**: Detected if `react-router.config.js` or
  `react-router.config.ts` exists, or if the build script contains
  `react-router build`. To enable SPA mode, set `ssr: false` in your React
  Router config.
- **Expo Web**: Detected if `expo` and `react-native-web` are in
  dependencies and `app.json` sets `expo.web.output` to `static` or `single`

For all frameworks, Railpack will try to detect the output directory and will
default to `dist` (or `build/client/` for React Router, or `out` for Next.js
static exports). Next.js reads `distDir` from your config when set. Set the
`RAILPACK_SPA_OUTPUT_DIR` environment variable to specify a custom output
directory. Railpack uses your app's `build` script to produce the static
output.

Note that if a SPA framework is *not* detected automatically, can you force SPA mode
by specifying a `RAILPACK_SPA_OUTPUT_DIR` environment variable. This will enable SPA
mode and serve the specified directory as a static site. Some of the SPA detection
uses regexes on framework configuration files, which will fail if the default framework
configuration files are customized.

Static sites are served using the [Caddy](https://caddyserver.com/) web server
and a [default
Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/node/Caddyfile.template).
You can overwrite this file with your own Caddyfile at the root of your project.

Node SPA deploys honor the `index_fallback` key in a `Staticfile` at the
project root when set. SPA routing behavior is unchanged unless you set
`index_fallback: false` (for example on multi-page Astro static sites) so
unknown paths return 404 and serve `404.html` when present.

## Framework Support

Railpack detects and configures caches and commands for popular frameworks.
Including:

- Next.js: Caches `.next/cache` for each Next.js app in the workspace
- Remix: Caches `.cache`
- Vite: Caches `node_modules/.vite`
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

Railpack automatically installs system dependencies for Puppeteer:

- **Puppeteer**: When detected in workspace dependencies, Railpack installs
  all necessary system packages for running headless Chrome, including
  `xvfb`, `chromium` dependencies, and font libraries. Note that
  Puppeteer's bundled Chromium [does not support
  ARM64](https://github.com/puppeteer/puppeteer/issues/7740); if you need
  to run on ARM hardware, consider switching to
  [Playwright](#playwright) or implementing a custom workaround
  (e.g. installing a system Chromium and pointing `executablePath` at it).
