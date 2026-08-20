---
title: Changelog
description: Release notes for each published version of Railpack.
editUrl: false
tableOfContents:
  minHeadingLevel: 2
  maxHeadingLevel: 2
---

## v0.37.0
August 18, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.37.0)

### Breaking Changes

* **Debian 13 base images:** New builds now use Debian 13 (Trixie), requiring custom Apt package lists and pinned base images to be compatible with Debian 13. ([#698]([#698](https://github.com/railwayapp/railpack/pull/698)#user-content-breaking-changes))
* **GCC 14 native builds:** Native gem and extension compilation now uses GCC 14, treating incompatible pointer types as errors and potentially requiring dependency updates or compiler flags. ([#705]([#705](https://github.com/railwayapp/railpack/pull/705)#user-content-breaking-changes))

### CLI

#### New

* **Debian 13:** Builder and runtime base images are now upgraded to Debian 13 (Trixie). Configure [custom Apt packages](https://railpack.com/guides/installing-packages#apt) with Debian 13 package names using `RAILPACK_DEPLOY_APT_PACKAGES` or `deploy.aptPackages`. by @iloveitaly in [#698](https://github.com/railwayapp/railpack/pull/698)

#### Fixed

* **Debian 13:** Builder images now include `libicu-dev` to ensure .NET restore and native gem compilation succeed on Debian 13, and Railpack warns during build planning when custom Apt packages are detected. by @iloveitaly in [#705](https://github.com/railwayapp/railpack/pull/705)

### Mise Upgrades

Updated mise from v2026.8.4 to [v2026.8.6](https://github.com/jdx/mise/releases/tag/v2026.8.6).

* **Resumable downloads:** Interrupted artifact downloads now resume via HTTP Range requests, and transient network errors such as HTTP/2 stream refusals are automatically retried. ([v2026.8.6](https://github.com/jdx/mise/releases/tag/v2026.8.6))
* **Precompiled PyPy:** PyPy can now be installed via precompiled binaries when compilation is disabled. ([v2026.8.5](https://github.com/jdx/mise/releases/tag/v2026.8.5))
* **Node source patching:** Node source builds can now apply local or remote patches before configuration using `node.apply_patches`. ([v2026.8.5](https://github.com/jdx/mise/releases/tag/v2026.8.5))
* **Rust nightly toolchains:** The rolling `nightly` channel resolves to concrete date-based toolchains for reproducible lockfiles and caching. ([v2026.8.6](https://github.com/jdx/mise/releases/tag/v2026.8.6))

**Full Changelog**: [v0.36.4...v0.37.0](https://github.com/railwayapp/railpack/compare/v0.36.4...v0.37.0)

## v0.36.4
August 12, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.36.4)

### What's Changed
* chore: mise update 2026.8.4 by @github-actions[bot] in [#689](https://github.com/railwayapp/railpack/pull/689)
* fix: exit 75 for transient failures by @edganiukov in [#690](https://github.com/railwayapp/railpack/pull/690)

### New Contributors
* @edganiukov made their first contribution in [#690](https://github.com/railwayapp/railpack/pull/690)

**Full Changelog**: [v0.36.3...v0.36.4](https://github.com/railwayapp/railpack/compare/v0.36.3...v0.36.4)

## v0.36.3
August 11, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.36.3)

### What's Changed
* Revert "fix: support multiline environment values and inheritance" by @coffee-cup in [#687](https://github.com/railwayapp/railpack/pull/687)
* fix: drop empty env entries instead of keeping explicit empty values by @coffee-cup in [#688](https://github.com/railwayapp/railpack/pull/688)

**Full Changelog**: [v0.36.1...v0.36.3](https://github.com/railwayapp/railpack/compare/v0.36.1...v0.36.3)

## v0.36.2
August 11, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.36.2)

**Full Changelog**: [v0.36.1...v0.36.2](https://github.com/railwayapp/railpack/compare/v0.36.1...v0.36.2)

## v0.36.1
August 11, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.36.1)

### What's Changed
* feat(build): add RAILPACK_BUILT_AT runtime variable by @iloveitaly in [#669](https://github.com/railwayapp/railpack/pull/669)
* feat(cli): add show-plan flag to info command by @iloveitaly in [#684](https://github.com/railwayapp/railpack/pull/684)
* fix: support multiline environment values and inheritance by @iloveitaly in [#685](https://github.com/railwayapp/railpack/pull/685)
* fix(mise): enable safe mode when reading app config by @coffee-cup in [#686](https://github.com/railwayapp/railpack/pull/686)

**Full Changelog**: [v0.36.0...v0.36.1](https://github.com/railwayapp/railpack/compare/v0.36.0...v0.36.1)

## v0.36.0
August 10, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.36.0)

### Providers

#### New

* **Node:** [TanStack Start](https://railpack.com/languages/node/#framework-support) apps are detected via `@tanstack/react-start` and are no longer treated as static Vite SPAs. If no `start` script is defined, Railpack installs `srvx` and starts with `srvx --prod -s ../client dist/server/server.js`; for production Node deploys, configure Nitro per TanStack's hosting guidance. by @iloveitaly in [#672](https://github.com/railwayapp/railpack/pull/672)
* **Node:** the default Node version is now `lts` instead of a fixed major, and non-semver aliases such as `lts` resolve correctly through mise. Pin a version with `RAILPACK_NODE_VERSION`, `engines.node`, or a version file as described in the [Node versions](https://railpack.com/languages/node/#versions) docs. by @iloveitaly in [#650](https://github.com/railwayapp/railpack/pull/650)
* **Node:** projects with a `package-lock.json` are detected as npm when no other package manager is configured. Prefer pinning Node and package manager versions for reproducible builds; see the [dependency recommendations](https://railpack.com/architecture/recommendations). by @iloveitaly in [#677](https://github.com/railwayapp/railpack/pull/677)
* **Rust:** [rust-toolchain.toml](https://railpack.com/languages/rust/#versions) is honored as an idiomatic version file (including channel, profile, components, and targets). Add a root `rust-toolchain.toml` such as `channel = "1.85.0"` and Railpack will install that toolchain via mise. by @iloveitaly in [#676](https://github.com/railwayapp/railpack/pull/676)
* **.NET:** [global.json](https://railpack.com/languages/dotnet/#versions) is honored as an idiomatic version file so `sdk.version` selects the .NET SDK. Add a root `global.json` such as `{ "sdk": { "version": "8.0.100" } }` to pin the SDK. by @iloveitaly in [#675](https://github.com/railwayapp/railpack/pull/675)

### CLI

#### New

* **Build context:** `.dockerignore` patterns (and top-level `exclude` in `railpack.json`) are applied when the local source is first loaded into BuildKit, so large ignored directories like `node_modules` are no longer transferred during local builds. Configure exclusions in `.dockerignore` or via top-level `exclude` in [railpack.json](https://railpack.com/config/excluding-files). by @iloveitaly in [#373](https://github.com/railwayapp/railpack/pull/373)

#### Fixed

* **Config:** explicit `deploy.inputs` with include filters are no longer overridden by an implicit full-step deploy output, preserving selections such as Next.js standalone output. by @radiantjade in [#562](https://github.com/railwayapp/railpack/pull/562)
* **Secrets:** CLI `--env` secrets are preserved when a config file also defines secrets, instead of being dropped. by @cadeljones in [#460](https://github.com/railwayapp/railpack/pull/460)

### Mise Upgrades

Updated mise from v2026.7.15 to [v2026.8.3](https://github.com/jdx/mise/releases/tag/v2026.8.3).

* **Precompiled Ruby:** when `ruby.compile` is unset, mise installs precompiled Ruby binaries by default and falls back to source builds only when no binary is available. ([v2026.8.0](https://github.com/jdx/mise/releases/tag/v2026.8.0))
* **Release-age filters:** `minimum_release_age` and related filters use GitHub's `published_at` timestamp so newly published releases of older commits no longer bypass age gates across Aqua, GitHub, Ubi, and related backends. ([v2026.8.3](https://github.com/jdx/mise/releases/tag/v2026.8.3))
* **npm lockfile installs:** packages pinned in `mise.lock` are trusted through the low-download popularity gate, so locked npm tool installs no longer require `allow_low_downloads`. ([v2026.7.16](https://github.com/jdx/mise/releases/tag/v2026.7.16))
* **Idiomatic version files:** disable individual idiomatic files per tool with `idiomatic_version_file_disable_files` (for example `node:package.json`) while keeping others active. ([v2026.7.17](https://github.com/jdx/mise/releases/tag/v2026.7.17))

**Full Changelog**: [v0.35.0...v0.36.0](https://github.com/railwayapp/railpack/compare/v0.35.0...v0.36.0)

## v0.35.0
July 28, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.35.0)

### Providers

#### New

* **Node & Python:** Playwright browser installation is now opt-in, preventing unexpected image-size increases. Set `RAILPACK_NODE_PLAYWRIGHT_INSTALL=1` or `RAILPACK_PYTHON_PLAYWRIGHT_INSTALL=1` when Playwright is a production dependency; see the [Node](https://railpack.com/languages/node/#playwright) and [Python](https://railpack.com/languages/python/#playwright) documentation. by @iloveitaly in [#653](https://github.com/railwayapp/railpack/pull/653)

### CLI

#### New

* **Mise:** Updated mise version from v2026.7.14 to [v2026.7.15](https://github.com/jdx/mise/releases/tag/v2026.7.15).

**Full Changelog**: [v0.34.0...v0.35.0](https://github.com/railwayapp/railpack/compare/v0.34.0...v0.35.0)

## v0.34.0
July 27, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.34.0)

### Providers

#### New

* **Node.js:** npm projects now use `npm install` by default, preventing builds from failing when `package-lock.json` is missing or out of sync. Set `RAILPACK_NODE_NPM_INSTALL="npm ci"` to opt into deterministic installs; Railpack also suggests committing missing npm and pnpm lockfiles, as described in the [dependency recommendations](https://railpack.com/architecture/recommendations). by @iloveitaly in [#643](https://github.com/railwayapp/railpack/pull/643)

### CLI

#### New

* **Mise:** Updated mise version from v2026.7.11 to [v2026.7.14](https://github.com/jdx/mise/releases/tag/v2026.7.14).

**Full Changelog**: [v0.33.0...v0.34.0](https://github.com/railwayapp/railpack/compare/v0.33.0...v0.34.0)

## v0.33.0
July 23, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.33.0)

### Providers

#### New

* **SvelteKit:** Projects using `@sveltejs/adapter-auto` now automatically produce and run a Node server without manual adapter configuration. Keep `@sveltejs/adapter-auto` in your dependencies, and Railpack will select `adapter-node` during the build and start it with `node build`. by @iloveitaly in [#641](https://github.com/railwayapp/railpack/pull/641)

#### Fixed

* **React Router:** [React Router SPA projects](https://railpack.com/languages/node#static-sites) with `ssr: false` are now correctly detected and served as static sites, even when the default `react-router-serve ./build/server/index.js` start script is present. by @iloveitaly in [#604](https://github.com/railwayapp/railpack/pull/604)

### CLI

#### New

* **Apt Packages:** Custom Apt package lists can now use the `...` spread operator to retain packages generated by Railpack. Follow the [package installation guide](https://railpack.com/guides/installing-packages#apt) and configure `"buildAptPackages": ["...", "build-essential"]` or `"deploy": { "aptPackages": ["...", "ffmpeg"] }`; omit `...` when you intentionally want to replace the generated list. by @iloveitaly in [#645](https://github.com/railwayapp/railpack/pull/645)

#### Fixed

* **Mise:** Railpack now uses musl-compatible mise binaries on Linux, fixing execution on Alpine and other musl-based hosts. by @iloveitaly in [#642](https://github.com/railwayapp/railpack/pull/642)

**Full Changelog**: [v0.32.0...v0.33.0](https://github.com/railwayapp/railpack/compare/v0.32.0...v0.33.0)

## v0.32.0
July 22, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.32.0)

### Providers

#### New

* **Node:** Railpack now supports [Next.js applications in Nx workspaces](https://railpack.com/languages/node#nx), including projects that use inferred tasks instead of root build and start scripts. Single-app workspaces are selected automatically; for multi-app workspaces, select an app with `RAILPACK_NX_APP=web railpack build .`. by @iloveitaly in [#606](https://github.com/railwayapp/railpack/pull/606)

#### Fixed

* **Node:** Railpack now displays a build-plan warning when no package manager is detected and npm is selected by default. by @iloveitaly in [#638](https://github.com/railwayapp/railpack/pull/638)

### CLI

#### New

* **Mise:** Updated mise version from v2026.7.7 to [v2026.7.11](https://github.com/jdx/mise/releases/tag/v2026.7.11).
* **Build Cache:** The [build command](https://railpack.com/reference/cli#build) now supports `--no-cache` for rebuilding without cached layers while preserving package-manager cache mounts. Run `railpack build --no-cache .` to perform a fresh build. by @iloveitaly in [#516](https://github.com/railwayapp/railpack/pull/516)
* **Build Output:** Build progress and completion output is now easier to scan, with a distinct start header and a formatted summary containing the build duration and next command. by @iloveitaly in [#640](https://github.com/railwayapp/railpack/pull/640)
* **Runtime Images:** Runtime images now expose [`RAILPACK_VERSION`](https://railpack.com/architecture/secrets#environment-variables), allowing applications and diagnostics to identify the Railpack version that produced the image. Access it through the `RAILPACK_VERSION` environment variable inside the running container. by @iloveitaly in [#633](https://github.com/railwayapp/railpack/pull/633)

**Full Changelog**: [v0.31.2...v0.32.0](https://github.com/railwayapp/railpack/compare/v0.31.2...v0.32.0)

## v0.31.2
July 20, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.31.2)

### Providers

#### Fixed

* **PHP:** PHP applications using a single Mise-managed runtime package now include it in the deploy image. by @jordyvandomselaar in [#635](https://github.com/railwayapp/railpack/pull/635)

### CLI

#### New

* **Mise:** Updated mise version from v2026.7.6 to [v2026.7.7](https://github.com/jdx/mise/releases/tag/v2026.7.7).

**Full Changelog**: [v0.31.1...v0.31.2](https://github.com/railwayapp/railpack/compare/v0.31.1...v0.31.2)

## v0.31.1
July 15, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.31.1)

### Providers

#### Fixed

* **Python:** Poetry is now installed via mise's native poetry package instead of pipx, improving environment integration. by @iloveitaly in [#625](https://github.com/railwayapp/railpack/pull/625)

### CLI

#### New

* **Mise:** Updated mise version from v2026.7.5 to [v2026.7.6](https://github.com/jdx/mise/releases/tag/v2026.7.6).

**Full Changelog**: [v0.31.0...v0.31.1](https://github.com/railwayapp/railpack/compare/v0.31.0...v0.31.1)

## v0.31.0
July 13, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.31.0)

### CLI

#### New

* **Build:** The [`build` command](https://railpack.com/reference/cli/#build) now supports BuildKit [cache import and export](https://railpack.com/architecture/caching/#cache-backends) via `--cache-from` and `--cache-to`, using the same syntax as docker buildx. For example: `railpack build --cache-from type=registry,ref=my.registry/cache --cache-to type=registry,ref=my.registry/cache,mode=max .` by @mvanhorn in [#595](https://github.com/railwayapp/railpack/pull/595)
* **Build:** `railpack build` now uses credentials from your Docker CLI config (`$DOCKER_CONFIG`, default `~/.docker/config.json`) so BuildKit can pull from and push to private registries. Log in with `docker login` first if needed. by @iloveitaly in [#623](https://github.com/railwayapp/railpack/pull/623)
* **Mise:** Updated mise version from v2026.7.2 to [v2026.7.5](https://github.com/jdx/mise/releases/tag/v2026.7.5).

**Full Changelog**: [v0.30.1...v0.31.0](https://github.com/railwayapp/railpack/compare/v0.30.1...v0.31.0)

## v0.30.1
July 9, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.30.1)

### CLI

#### New

* **Mise:** Updated mise version from v2026.6.12 to [v2026.7.2](https://github.com/jdx/mise/releases/tag/v2026.7.2).

**Full Changelog**: [v0.30.0...v0.30.1](https://github.com/railwayapp/railpack/compare/v0.30.0...v0.30.1)

## v0.30.0
June 22, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.30.0)

### Providers

#### New

* **Deno:** Railpack now reads Deno versions from `.deno-version` files using mise idiomatic version file support. Add a `.deno-version` file to your project root (for example, `2.2.2`) to pin the Deno runtime for your deployment — see [version detection](https://railpack.com/languages/deno#versions) for the full precedence order. by @iloveitaly in [#608](https://github.com/railwayapp/railpack/pull/608)

#### Fixed

* **Python, Ruby, Deno:** `RAILPACK_PYTHON_VERSION`, `RAILPACK_RUBY_VERSION`, and `RAILPACK_DENO_VERSION` environment variables now correctly override local version files and mise configuration during builds, matching the documented precedence order. Set these in your deployment environment when you need to pin a runtime version independently of repository config. by @Vunas in [#610](https://github.com/railwayapp/railpack/pull/610)

### CLI

#### New

* **Mise:** Updated mise version from v2026.6.11 to [v2026.6.12](https://github.com/jdx/mise/releases/tag/v2026.6.12).

**Full Changelog**: [v0.29.0...v0.30.0](https://github.com/railwayapp/railpack/compare/v0.29.0...v0.30.0)

## v0.29.0
June 18, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.29.0)

### Providers

#### New

* **Node:** Railpack now detects and deploys Next.js apps configured for static export as SPAs, serving the built site with Caddy. Set `output: 'export'` in your `next.config.ts` (or `.js`/`.mjs`) and Railpack will build and serve the static output from `out` (or your configured `distDir`). See the [Node.js SPA documentation](https://railpack.com/languages/node#spa-mode) for details. by @iloveitaly in [#602](https://github.com/railwayapp/railpack/pull/602)

* **Node:** Node SPA deployments now honor the `index_fallback` key in a `Staticfile` at your project root. SPA routing is unchanged by default, but you can set `index_fallback: false` to disable `index.html` fallback on unknown routes—useful for multi-page static sites that need custom 404 pages. See the [Node.js SPA documentation](https://railpack.com/languages/node#spa-mode) for details. by @iloveitaly in [#582](https://github.com/railwayapp/railpack/pull/582)

* **Node:** Bun-only apps with no dependencies no longer install Node.js in the final image when it isn't needed. by @iloveitaly in [#601](https://github.com/railwayapp/railpack/pull/601)

#### Fixed

* **Node:** Expo projects using the default `expo start` command are no longer incorrectly treated as having a custom start command, so SPA detection works as expected. by @iloveitaly in [#582](https://github.com/railwayapp/railpack/pull/582)

* **Node:** `RAILPACK_NODE_VERSION` now correctly takes precedence over `engines.node` in `package.json`, matching the documented version resolution order. by @iloveitaly in [#601](https://github.com/railwayapp/railpack/pull/601)

**Full Changelog**: [v0.28.0...v0.29.0](https://github.com/railwayapp/railpack/compare/v0.28.0...v0.29.0)

## v0.28.0
June 17, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.28.0)

### Providers

#### New

* **Expo Web:** Railpack now detects [Expo Web](https://railpack.com/languages/node/) projects configured for static export as SPAs and serves the built output from `dist` with the static file server. Configure `expo.web.output` to `static` or `single` in `app.json`, ensure `expo` and `react-native-web` are in your dependencies, and add a `build` script such as `expo export --platform web` to your `package.json`. by @iloveitaly in [#585](https://github.com/railwayapp/railpack/pull/585)

#### Fixed

* **Node:** Increased the default npm fetch retry count to 5 (`NPM_CONFIG_FETCH_RETRIES=5`) to improve build reliability against transient registry network errors during dependency installation. by @iloveitaly in [#598](https://github.com/railwayapp/railpack/pull/598)

### CLI

#### New

* **Mise:** Updated mise version from v2026.6.10 to [v2026.6.11](https://github.com/jdx/mise/releases/tag/v2026.6.11).

**Full Changelog**: [v0.27.2...v0.28.0](https://github.com/railwayapp/railpack/compare/v0.27.2...v0.28.0)

## v0.27.2
June 16, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.27.2)

### CLI

#### New

* **Mise:** Updated mise version from v2026.6.5 to [v2026.6.10](https://github.com/jdx/mise/releases/tag/v2026.6.10).

**Full Changelog**: [v0.27.1...v0.27.2](https://github.com/railwayapp/railpack/compare/v0.27.1...v0.27.2)

## v0.27.1
June 13, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.27.1)

update mise to 2026.6.5

## v0.27.0
June 9, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.27.0)

### Providers

#### Fixed

* **Elixir:** Elixir 1.20 now maps to Erlang 29, so projects on Elixir 1.20 build with the correct OTP version. by @iloveitaly in [#583](https://github.com/railwayapp/railpack/pull/583)

### CLI

#### New

* **Mise:** Updated mise version from v2026.6.0 to [v2026.6.1](https://github.com/jdx/mise/releases/tag/v2026.6.1).

### Runtime

* **Locale:** Builder and runtime images now include the `en_US.UTF-8` locale, fixing warnings and breakage for apps that depend on locale-aware behavior (such as Python or Ruby). Railpack does not set `LANG` or `LC_ALL` by default — enable UTF-8 by adding them to [`deploy.variables`](https://railpack.com/config/file) in your `railpack.json` (for example, `"LANG": "en_US.UTF-8"`). by @iloveitaly in [#576](https://github.com/railwayapp/railpack/pull/576)
* **Timezone:** Builder and runtime images now include `tzdata` by default, so named timezones work out of the box when you set `TZ` or use libraries that read the zoneinfo database. by @iloveitaly in [#580](https://github.com/railwayapp/railpack/pull/580)

**Full Changelog**: [v0.26.1...v0.27.0](https://github.com/railwayapp/railpack/compare/v0.26.1...v0.27.0)

## v0.26.1
June 4, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.26.1)

### Providers

#### Fixed

* **Node:** pnpm 11+ projects now build correctly by adding `PNPM_HOME/bin` to the `PATH` so the pre-install `node-gyp` step can be found, including when `engines.pnpm` uses x-range notation like `"11.5.x"`. by @iloveitaly in [#581](https://github.com/railwayapp/railpack/pull/581)

### CLI

#### New

* **Mise:** Updated mise version from v2026.5.18 to [v2026.6.0](https://github.com/jdx/mise/releases/tag/v2026.6.0).

**Full Changelog**: [v0.26.0...v0.26.1](https://github.com/railwayapp/railpack/compare/v0.26.0...v0.26.1)

## v0.26.0
June 2, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.26.0)

### Providers

#### Fixed

* **Node (pnpm):** Railpack now sets `PNPM_HOME` to `/opt/pnpm` and `PNPM_STORE_DIR` to `/opt/pnpm/store` during every pnpm install so BuildKit caches the store path pnpm actually uses. This fixes missed install cache hits when newer pnpm versions write outside the old hardcoded store location. by @iloveitaly in [#569](https://github.com/railwayapp/railpack/pull/569)

### CLI

#### Fixed

* **Config:** Custom runtime images set via [`deploy.base`](https://railpack.com/config/file) in `railpack.json` are now applied to the generated deploy step. Previously Railpack parsed `deploy.base` but did not use it when building the final image; configured bases are used as the deploy filesystem and as the input for runtime apt package layers when those are enabled. by @radiantjade in [#561](https://github.com/railwayapp/railpack/pull/561)

* **Mise:** Builds no longer fail when `mise.lock` pins a tool version newer than the default [14-day minimum release age](https://railpack.com/config/mise) while your config requests `latest`. Host-side version resolution applies the age filter first, then falls back when a pinned lockfile version would otherwise be excluded. by @iloveitaly in [#575](https://github.com/railwayapp/railpack/pull/575)

**Full Changelog**: [v0.25.0...v0.26.0](https://github.com/railwayapp/railpack/compare/v0.25.0...v0.26.0)

## v0.25.0
June 1, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.25.0)

### Providers

#### New

* **Node:** Railpack now detects and builds projects that use `package.json5` manifests (in addition to `package.json`), including workspace and dependency parsing. If both files exist, `package.json` takes precedence. by @iloveitaly in [#566](https://github.com/railwayapp/railpack/pull/566)

* **Node:** Node 25+ apps that use Corepack (for example `packageManager: "pnpm@..."`) now include the mise install and shim directories in the runtime image, so package managers like `pnpm` remain available when your start command runs after the build. by @iloveitaly in [#568](https://github.com/railwayapp/railpack/pull/568)

#### Fixed

* **Node:** `engines.pnpm` in your manifest now takes precedence over pnpm version inferred from `pnpm-lock.yaml`, avoiding ambiguous lockfile version mapping across major pnpm releases. See [package and version resolution](https://railpack.com/architecture/package-resolution/) for how Railpack resolves tool versions. by @iloveitaly in [#566](https://github.com/railwayapp/railpack/pull/566)

### CLI

#### New

* **Config:** The `provider` field in `railpack.json` is now listed in the generated JSON schema (`railpack schema`), with the same values documented in the [config file guide](https://railpack.com/config/file/). Provider names are matched case-insensitively, so lowercase values like `elixir` and `dotnet` work as documented. by @radiantjade in [#558](https://github.com/railwayapp/railpack/pull/558)

* **Mise:** Updated mise version from v2026.5.16 to [v2026.5.18](https://github.com/jdx/mise/releases/tag/v2026.5.18).

**Full Changelog**: [v0.24.0...v0.25.0](https://github.com/railwayapp/railpack/compare/v0.24.0...v0.25.0)

## v0.24.0
May 29, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.24.0)

### Providers

#### New

* **Node:** Package manager versions for pnpm, bun, and yarn defined in mise config (`mise.toml`, `.tool-versions`, etc.) now take priority over versions inferred from lockfiles or project metadata. Pin a version in your config—for example `pnpm = "10.4.1"` in `mise.toml`—and Railpack will use it consistently across the build. See [Node.js versions](https://railpack.com/languages/node#versions) and [Mise configuration](https://railpack.com/config/mise). by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

#### Fixed

* **Node:** Builds using pnpm 11+ now add `PNPM_HOME/bin` to `PATH`, matching pnpm’s new layout so global tools like `node-gyp` install and run correctly. by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

* **Ruby:** Ruby source builds no longer install rdoc, avoiding locale-related compile failures during `mise install`. by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

* **Elixir:** Elixir provider environment variables (locale, `MIX_ENV`, `MIX_HOME`, etc.) are now applied to the mise install step so Erlang and Elixir install with the same settings as the rest of the build. by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

* **Python:** Projects using Poetry, PDM, or Pipenv via pipx now also install `uv`, which newer mise requires when `pipx.uvx` is enabled. by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

### CLI

#### New

* **Mise:** Updated mise version from v2026.3.17 to [v2026.5.16](https://github.com/jdx/mise/releases/tag/v2026.5.16).

#### Fixed

* **Mise:** Generated `/etc/mise/config.toml` now sets `minimum_release_age` instead of the renamed `install_before` setting, matching current mise behavior. Override it in your own [mise configuration](https://railpack.com/config/mise) if needed. by @iloveitaly in [#564](https://github.com/railwayapp/railpack/pull/564)

**Full Changelog**: [v0.23.0...v0.24.0](https://github.com/railwayapp/railpack/compare/v0.23.0...v0.24.0)

## v0.23.0
March 30, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.23.0)

### Providers

#### New

* **Staticfile:** Added a [configurable index fallback](https://railpack.com/languages/staticfile) to the Staticfile provider, allowing users to toggle SPA-style routing for their applications. by @iloveitaly in [#534](https://github.com/railwayapp/railpack/pull/534)
  To enable SPA routing, add `index_fallback: true` to your `Staticfile` configuration to have non-existent paths served by `index.html`.

### CLI

#### New

* **Mise:** Updated mise version from v2026.3.15 to [v2026.3.17](https://github.com/jdx/mise/releases/tag/v2026.3.17). in [#539](https://github.com/railwayapp/railpack/pull/539)
* **Mise:** Introduced a [new installation policy](https://railpack.com/config/mise) that defaults to tool versions released at least 14 days ago to avoid broken or incomplete tool versions. by @iloveitaly in [#540](https://github.com/railwayapp/railpack/pull/540)
  This setting improves stability for runtime environments like Python and Ruby by allowing upstream releases time to stabilize before they are used in production builds.
* **Verbose Mode:** Added support for the `RAILPACK_VERBOSE` environment variable to enable [verbose logging](https://railpack.com/config/environment-variables) and improved debug logs for mise execution. by @iloveitaly in [#536](https://github.com/railwayapp/railpack/pull/536)
  Set `RAILPACK_VERBOSE=1` to enable detailed debug output, including the full command string and captured output from mise during execution.

#### Fixed

* **Configuration:** Improved root directory detection by trimming whitespace from environment variables and configuration values. by @iloveitaly in [#534](https://github.com/railwayapp/railpack/pull/534)

**Full Changelog**: [v0.22.2...v0.23.0](https://github.com/railwayapp/railpack/compare/v0.22.2...v0.23.0)

## v0.22.2
March 26, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.22.2)

### CLI

#### New

* **Mise:** Updated mise version from v2026.3.13 to [v2026.3.15](https://github.com/jdx/mise/releases/tag/v2026.3.15). in [#532](https://github.com/railwayapp/railpack/pull/532)

**Full Changelog**: [v0.22.1...v0.22.2](https://github.com/railwayapp/railpack/compare/v0.22.1...v0.22.2)

## v0.22.1
March 24, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.22.1)

### CLI

#### New

* **Mise:** Updated mise version from v2026.3.12 to [v2026.3.13](https://github.com/jdx/mise/releases/tag/2026.3.13).

**Full Changelog**: [v0.22.0...v0.22.1](https://github.com/railwayapp/railpack/compare/v0.22.0...v0.22.1)

## v0.22.0
March 23, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.22.0)

### CLI

#### New

* **Configuration:** Railpack now uses a global `mise.toml` to store default tool configurations instead of environment variables, allowing users to naturally override defaults via their project's local `mise.toml` file. This change makes it easier to customize tool versions and settings without needing to manage complex environment variables. For example, you can now add a `[settings]` section to your project's `mise.toml` to specify exact settings that Railpack should use. by @iloveitaly in [#504](https://github.com/railwayapp/railpack/pull/504)
* **Mise:** Updated mise version from v2026.3.10 to [v2026.3.12](https://github.com/jdx/mise/releases/tag/2026.3.12).

**Full Changelog**: [v0.21.0...v0.22.0](https://github.com/railwayapp/railpack/compare/v0.21.0...v0.22.0)

## v0.21.0
March 21, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.21.0)

### What's Changed
* fix: expose shell completion subcommand by @iloveitaly in [#520](https://github.com/railwayapp/railpack/pull/520)
* test: add arm builds by @iloveitaly in [#509](https://github.com/railwayapp/railpack/pull/509)
* feat: additional mise config and idiomatic version files enabling precompiled ruby support by @iloveitaly in [#487](https://github.com/railwayapp/railpack/pull/487)
* chore: mise update 2026.3.10 by @github-actions[bot] in [#522](https://github.com/railwayapp/railpack/pull/522)

**Full Changelog**: [v0.20.0...v0.21.0](https://github.com/railwayapp/railpack/compare/v0.20.0...v0.21.0)

## v0.20.0
March 18, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.20.0)

### Providers

### CLI

#### New

* **BuildKit:** Images and build cache are now exported in parallel, significantly reducing the time spent in the final stages of the build process. This feature utilizes the latest BuildKit v0.28.0 capabilities and is enabled automatically for all `railpack build` operations. by @iloveitaly in [#515](https://github.com/railwayapp/railpack/pull/515)
* **Shell:** Added native support for **Fish shell completion**. You can generate the completion script for your shell by running `railpack completion fish` and adding it to your fish configuration. by @iloveitaly in [#515](https://github.com/railwayapp/railpack/pull/515)
* toml 1.1 features are supported. This could have caused errors in the past if railpack read a toml file with previously unsupported features

#### Fixed

* **Mise:** Updated the internal environment variable to `MISE_SYSTEM_CONFIG_DIR` to ensure better compatibility and isolation from host configurations when running `railpack`. by @iloveitaly in [#517](https://github.com/railwayapp/railpack/pull/517)

**Full Changelog**: [v0.19.0...v0.20.0](https://github.com/railwayapp/railpack/compare/v0.19.0...v0.20.0)

## v0.19.0
March 16, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.19.0)

### Providers

#### New

* **Mise** Updated mise version from v2026.2.14 to [v2026.3.9](https://github.com/jdx/mise/releases/tag/v2026.3.9).
* **Node.js:** Node.js projects using Bun now support [bunfig.toml](https://railpack.com/languages/node) during the installation step. This ensures custom registries and install behaviors are correctly applied when running `bun install`. by @iloveitaly in [#477](https://github.com/railwayapp/railpack/pull/477)
* **Ruby:** Upgraded to [jemalloc2](https://railpack.com/languages/ruby) for Ruby projects to resolve errors on ARM architecture and improve memory performance. This is enabled by default. by @iloveitaly in [#338](https://github.com/railwayapp/railpack/pull/338)
* **Python:** Added documentation and an example for using [compiled Python](https://railpack.com/languages/python) as an escape hatch when pre-built binaries are not available for specific versions. This allows developers to use specific Python builds by opting into compilation during the build process. by @iloveitaly in [#503](https://github.com/railwayapp/railpack/pull/503)

#### Fixed

* **Ruby:** Default Ruby versions are now pinned to the minor version instead of a specific patch to allow for automatic security updates and better compatibility with pre-built binaries. by @iloveitaly in [#488](https://github.com/railwayapp/railpack/pull/488)
* **Staticfile:** Switched to a binary installer for Caddy to achieve [smaller image sizes](https://railpack.com/languages/staticfile) for static site deployments. by @iloveitaly in [#478](https://github.com/railwayapp/railpack/pull/478)

### CLI

#### Fixed

* **CLI:** The CLI now issues a warning if no start command is detected during the build process, providing clearer feedback on potentially missing configuration. by @iloveitaly in [#490](https://github.com/railwayapp/railpack/pull/490)

**Full Changelog**: [v0.18.0...v0.19.0](https://github.com/railwayapp/railpack/compare/v0.18.0...v0.19.0)

## v0.18.0
March 9, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.18.0)

This is mostly an internal release to ensure that the mise version used in the planning, building, and runtime images are all in sync.

The only user-facing change in this release in bumping the build and runtime mise version to `2026.2.22`.

**Full Changelog**: [v0.17.2...v0.18.0](https://github.com/railwayapp/railpack/compare/v0.17.2...v0.18.0)

## v0.17.2
February 12, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.17.2)

### What's Changed
* chore: mise update 2026.1.12 by @iloveitaly in [#451](https://github.com/railwayapp/railpack/pull/451)
* fix: eliminate additional logs when using --dump-llb by @iloveitaly in [#452](https://github.com/railwayapp/railpack/pull/452)
* feat(python): Scan nested dependency files in usesDep by @half0wl in [#457](https://github.com/railwayapp/railpack/pull/457)

### New Contributors
* @half0wl made their first contribution in [#457](https://github.com/railwayapp/railpack/pull/457)

**Full Changelog**: [v0.17.1...v0.17.2](https://github.com/railwayapp/railpack/compare/v0.17.1...v0.17.2)

## v0.17.1
January 19, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.17.1)

### What's Changed
* fix: revert playwright support by @iloveitaly in [#447](https://github.com/railwayapp/railpack/pull/447)

**Full Changelog**: [v0.17.0...v0.17.1](https://github.com/railwayapp/railpack/compare/v0.17.0...v0.17.1)

## v0.17.0
January 19, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.17.0)

### Providers

#### New

* **Node & Python:** We've added automatic [Playwright support](https://railpack.com/docs/languages/node) for Node.js and Python projects. Railpack now detects Playwright dependencies and automatically installs the required system packages and browsers. by @iloveitaly in [#348](https://github.com/railwayapp/railpack/pull/348)
* **Shell:** The shell provider now supports a `mise` install phase, allowing you to define tools to install before your script runs. by @iloveitaly in [#433](https://github.com/railwayapp/railpack/pull/433)

#### Fixed

* **Node:** Consistent Node version resolution is now used when Bun is the package manager. by @iloveitaly in [#359](https://github.com/railwayapp/railpack/pull/359)
* **Node:** Trailing whitespace in the `packageManager` field of `package.json` is now handled correctly. by @iloveitaly in [#432](https://github.com/railwayapp/railpack/pull/432)
* **Node:** Improved detection of version files (like `.nvmrc`) using Mise's idiomatic parsing. by @iloveitaly in [#431](https://github.com/railwayapp/railpack/pull/431)
* **Python:** Builds are now faster for projects using `psycopg-binary` as unnecessary system dependencies (libpq-dev) are skipped. by @iloveitaly in [#379](https://github.com/railwayapp/railpack/pull/379)
* **Python:** `.python-version` files are now reliably detected in all environments. by @iloveitaly in [#378](https://github.com/railwayapp/railpack/pull/378)

### CLI

#### Fixed

* **Dockerignore:** Negation rules (e.g., `!file`) in `.dockerignore` now work correctly even if the referenced file does not exist locally. by @iloveitaly in [#437](https://github.com/railwayapp/railpack/pull/437)
* **Config:** The `provider` field is no longer required in `railpack.json`. by @iloveitaly in [#436](https://github.com/railwayapp/railpack/pull/436)

**Full Changelog**: [v0.16.0...v0.17.0](https://github.com/railwayapp/railpack/compare/v0.16.0...v0.17.0)

## v0.16.0
January 14, 2026 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.16.0)

### Providers

#### New

* **Python:** Projects with `start.py` or `bot.py` are now automatically detected as Python apps. This simplifies configuration for simple scripts and bots (e.g. Discord bot
that don't have manifest files. by @coffee-cup in [#416](https://github.com/railwayapp/railpack/pull/416)

#### Fixed

* **Dotnet:** The web server now correctly listens on the port defined by the `PORT` environment variable. See the [updated documentation](
https://railpack.com/languages/dotnet). by @iloveitaly in [#401](https://github.com/railwayapp/railpack/pull/401)

### CLI

#### Fixed

* **Build:** Generated Docker image names are now forced to lowercase to prevent build errors with registries that require lowercase repositories. by @iloveitaly in
[#410](https://github.com/railwayapp/railpack/pull/410)

**Full Changelog**: [v0.15.3...v0.15.4](https://github.com/railwayapp/railpack/compare/v0.15.3...v0.15.4)

## v0.15.4
December 20, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.15.4)

### Providers

#### Fixed

* **Dotnet:** set aspnet content root to properly find configuration files by @iloveitaly and @Mohs9n in [#396](https://github.com/railwayapp/railpack/pull/396)
* **Erlang:** There is an underlying bug in mise that causes the version selection to break. This has been fixed within Railpack and an upstream change has been submitted to mise.

**Full Changelog**: [v0.15.2...v0.15.3](https://github.com/railwayapp/railpack/compare/v0.15.2...v0.15.3)

## v0.15.3
December 19, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.15.3)

### What's Changed
* sync mise version between Go code and Dockerfile by @coffee-cup in [#394](https://github.com/railwayapp/railpack/pull/394)

**Full Changelog**: [v0.15.2...v0.15.3](https://github.com/railwayapp/railpack/compare/v0.15.2...v0.15.3)

## v0.15.2
December 19, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.15.2)

### What's Changed
* update mise veresions by @coffee-cup in [#389](https://github.com/railwayapp/railpack/pull/389)
* build mise image faster by @coffee-cup in [#390](https://github.com/railwayapp/railpack/pull/390)
* release build mise for now by @coffee-cup in [#391](https://github.com/railwayapp/railpack/pull/391)
* install mise from script by @coffee-cup in [#392](https://github.com/railwayapp/railpack/pull/392)
* fix builder mise location by @coffee-cup in [#393](https://github.com/railwayapp/railpack/pull/393)

**Full Changelog**: [v0.15.1...v0.15.2](https://github.com/railwayapp/railpack/compare/v0.15.1...v0.15.2)

## v0.15.1
December 1, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.15.1)

### What's Changed
* perf: optimize plan command for large monorepos by @coffee-cup in [#382](https://github.com/railwayapp/railpack/pull/382)
* fix: deduplicate deploy inputs when using --build-cmd by @coffee-cup in [#381](https://github.com/railwayapp/railpack/pull/381)

**Full Changelog**: [v0.15.0...v0.15.1](https://github.com/railwayapp/railpack/compare/v0.15.0...v0.15.1)

## v0.15.0
November 29, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.15.0)

### What's Changed
* chore: mise update 2025.11.8 by @iloveitaly in [#377](https://github.com/railwayapp/railpack/pull/377)
* fix: consider exclude patterns when determining LLB merge eligibility by @coffee-cup in [#380](https://github.com/railwayapp/railpack/pull/380)

**Full Changelog**: [v0.14.0...v0.15.0](https://github.com/railwayapp/railpack/compare/v0.14.0...v0.15.0)

## v0.14.0
November 24, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.14.0)

### Providers

#### New

*   **C/C++:** We've added a [C/C++ provider](https://railpack.com/languages/cpp) supporting CMake and Meson build systems. by @aleksrutins in [#319](https://github.com/railwayapp/railpack/pull/319)
*   **Python:** We've added automatic start command detection for [FastAPI](https://railpack.com/languages/python) applications using `uvicorn`. by @iloveitaly in [#353](https://github.com/railwayapp/railpack/pull/353)

#### Fixed

*   **Python:** We've fixed detection of `psycopg` (v3) to ensure required system dependencies (libpq) are installed. by @iloveitaly in [#350](https://github.com/railwayapp/railpack/pull/350)
*   **General:** We've replaced fuzzy file matching with exact checks to prevent false positives when detecting providers. by @iloveitaly in [#347](https://github.com/railwayapp/railpack/pull/347)

### CLI

#### Fixed

*   **General:** `mise` configuration files in the application directory are now automatically trusted, suppressing "untrusted config" warnings. by @iloveitaly in [#352](https://github.com/railwayapp/railpack/pull/352)

**Full Changelog**: [v0.13.0...v0.14.0](https://github.com/railwayapp/railpack/compare/v0.13.0...v0.14.0)

## v0.13.0
November 18, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.13.0)

### Providers

#### New

* **Node:** Added support for Node.js v25 by ensuring `libatomic1` is explicitly installed. by @iloveitaly in [#259](https://github.com/railwayapp/railpack/pull/259)
* **Node:** Railpack now ensures Node.js is available during the build phase via mise even when not required for the runtime (e.g. when using Bun), ensuring packages requiring `node-gyp` build correctly. by @iloveitaly in [#259](https://github.com/railwayapp/railpack/pull/259)

#### Fixed

* **Node:** We've improved how `node_modules` caching works during the build phase, specifically fixing issues where commands like `npm ci` (which remove `node_modules`) caused build failures. See the [Node documentation](https://railpack.com/languages/node) for more details. by @iloveitaly in [#259](https://github.com/railwayapp/railpack/pull/259)

### CLI

#### New

* **Mise:** Updated the internal `mise` version to 2025.11.6. by @iloveitaly in https://github.com/railwayapp/railpack/commit/d249231615dda1f2b4d7ade419f07890e2896079

#### Fixed

* **Mise:** Added backoff retries for `mise` registry downloads to help prevent build failures during intermittent network issues. by @iloveitaly in [#360](https://github.com/railwayapp/railpack/pull/360)

**Full Changelog**: [v0.12.0...v0.13.0](https://github.com/railwayapp/railpack/compare/v0.12.0...v0.13.0)

## v0.12.0
November 17, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.12.0)

### Providers

#### New

* **Go:** The Go provider now respects the Go version specified in your `mise` configuration files (e.g., `mise.toml`, `.tool-versions`). by @iloveitaly in [#300](https://github.com/railwayapp/railpack/pull/300)
* **Ruby:** The Ruby provider now respects the Ruby version specified in your `mise` configuration files (e.g., `mise.toml`, `.tool-versions`). by @iloveitaly in [#339](https://github.com/railwayapp/railpack/pull/339)
* **Node:** For Bun projects that require Node.js for build tooling (like `node-gyp`), Railpack now installs Node.js using `mise`, respecting versions from `.nvmrc`, `.node-version`, and other `mise` configurations. by @iloveitaly in [#341](https://github.com/railwayapp/railpack/pull/341)
* **Gleam:** You can now include your project's source code in the final image by setting the `RAILPACK_GLEAM_INCLUDE_SOURCE=true` environment variable. by @iloveitaly in [#318](https://github.com/railwayapp/railpack/pull/318)

#### Fixed

* **Python:** The Python provider now configures `pipx` to use `uv` for installing packages, resulting in faster dependency installation. by @iloveitaly in [#324](https://github.com/railwayapp/railpack/pull/324)
* **All:** Railpack now parses `mise` configuration files (like `mise.toml` and `.tool-versions`) more idiomatically, improving compatibility and correctness. by @iloveitaly in [#344](https://github.com/railwayapp/railpack/pull/344)

### CLI

#### New

* **Caching:** You can now disable specific caches during the build process by setting the `RAILPACK_DISABLE_CACHE` environment variable. For example, to disable the `apt` and `pip` caches, set `RAILPACK_DISABLE_CACHE="apt,pip"`. To disable all caches, set it to `*`. by @aleksrutins in [#316](https://github.com/railwayapp/railpack/pull/316)

#### Fixed

* **Security:** Railpack now enables `mise`'s "paranoid mode" by default, a security feature that requires explicit user consent before installing new tools or running code from project-specific configurations. by @iloveitaly in [#336](https://github.com/railwayapp/railpack/pull/336)

**Full Changelog**: [v0.11.0...v0.12.0](https://github.com/railwayapp/railpack/compare/v0.11.0...v0.12.0)

## v0.11.0
November 11, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.11.0)

### Providers

#### New

*   **Dotnet:** Railpack now supports building [.NET applications](https://railpack.com/languages/dotnet). The provider defaults to the latest LTS version of .NET, and the version can be customized using `mise` configuration files (e.g., `.tool-versions`, `mise.toml`). by @jaredLunde and @iloveitaly in [#126](https://github.com/railwayapp/railpack/pull/126) and [#323](https://github.com/railwayapp/railpack/pull/323)
*   **Shell:** The [Shell provider](https://railpack.com/languages/shell) now supports `zsh` as an interpreter and automatically detects the required shell (bash or zsh) from your start script's shebang (e.g., `#!/bin/zsh`). by @iloveitaly in [#328](https://github.com/railwayapp/railpack/pull/328) and [#327](https://github.com/railwayapp/railpack/pull/327)
*   **Deno:** You can now specify the Deno version for your project using `mise` configuration files like `.tool-versions` or `mise.toml` with the [Deno provider](https://railpack.com/languages/deno). by @iloveitaly in [#325](https://github.com/railwayapp/railpack/pull/325)

**Full Changelog**: [v0.10.0...v0.11.0](https://github.com/railwayapp/railpack/compare/v0.10.0...v0.11.0)

## v0.10.0
November 6, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.10.0)

### Providers

#### New

* **Gleam:** Railpack now has first-class support for Gleam projects. by @aleksrutins in [#313](https://github.com/railwayapp/railpack/pull/313)
* **Elixir:** Erlang and Elixir versions defined in `mise` config files (e.g., `mise.toml`, `.tool-versions`) will now be used instead of Railpack's defaults. by @iloveitaly in [#306](https://github.com/railwayapp/railpack/pull/306)
* **Node.js:** The `engines` field in `package.json` is now used to determine the versions of Node.js and the package manager (npm, yarn, pnpm), but only if a version is not specified in a lockfile or other config. by @iloveitaly in [#251](https://github.com/railwayapp/railpack/pull/251)
  For example, to use Node.js v20 and pnpm v9, you can add the following to your `package.json`:
  ```json
  {
    "engines": {
      "node": "20.x",
      "pnpm": "9.x"
    }
  }
  ```

#### Fixed

* **Python:** `mise` Python compilation is disabled by default to avoid package compilation issues on newly released python versions. by @coffee-cup in [#301](https://github.com/railwayapp/railpack/pull/301)

**Full Changelog**: [v0.9.2...v0.10.0](https://github.com/railwayapp/railpack/compare/v0.9.2...v0.10.0)

## v0.9.2
October 22, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.9.2)

### What's Changed
* chore: mise update 2025.10.11 by @iloveitaly in [#297](https://github.com/railwayapp/railpack/pull/297)
* docs: fix env var display on docs by @iloveitaly in [#298](https://github.com/railwayapp/railpack/pull/298)
* docs: update python with mise-compatible version files by @crisog in [#290](https://github.com/railwayapp/railpack/pull/290)
* fix: use default mise backend for uv by @iloveitaly in [#260](https://github.com/railwayapp/railpack/pull/260)
* test: http checks on example projects by @iloveitaly in [#250](https://github.com/railwayapp/railpack/pull/250)
* chore: mise update 2025.10.13 by @iloveitaly in [#302](https://github.com/railwayapp/railpack/pull/302)
* build: add mise upgrade task by @iloveitaly in [#303](https://github.com/railwayapp/railpack/pull/303)
* test: assert against specific rust version defined by rust-toolchain.toml by @iloveitaly in [#299](https://github.com/railwayapp/railpack/pull/299)
* fix: use channel key not version from rust-toolchain.toml by @iloveitaly in [#304](https://github.com/railwayapp/railpack/pull/304)
* refactor: cleanup rust provider by @iloveitaly in [#305](https://github.com/railwayapp/railpack/pull/305)
* docs: minor docs and comments by @iloveitaly in [#309](https://github.com/railwayapp/railpack/pull/309)

### New Contributors
* @crisog made their first contribution in [#290](https://github.com/railwayapp/railpack/pull/290)

**Full Changelog**: [v0.9.1...v0.9.2](https://github.com/railwayapp/railpack/compare/v0.9.1...v0.9.2)

## v0.9.1
October 16, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.9.1)

### What's Changed
* fix: support + character in environment variable names by @coffee-cup in [#294](https://github.com/railwayapp/railpack/pull/294)

**Full Changelog**: [v0.9.0...v0.9.1](https://github.com/railwayapp/railpack/compare/v0.9.0...v0.9.1)

## v0.9.0
October 7, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.9.0)

### What's Changed
* fix: use containerd platform parsing to fix platform parsing by @iloveitaly in [#277](https://github.com/railwayapp/railpack/pull/277)
* test: php snapshot update by @iloveitaly in [#280](https://github.com/railwayapp/railpack/pull/280)
* chore: mise update 2025.9.19 by @iloveitaly in [#281](https://github.com/railwayapp/railpack/pull/281)
* feat: add --hide-pretty-plan and --show-plan to prepare by @iloveitaly in [#284](https://github.com/railwayapp/railpack/pull/284)
* feat: use py package versions from any mise-supported version specification file by @iloveitaly in [#261](https://github.com/railwayapp/railpack/pull/261)
* chore: mise update 2025.9.25 by @iloveitaly in [#288](https://github.com/railwayapp/railpack/pull/288)
* chore: mise update 2025.10.4 by @iloveitaly in [#289](https://github.com/railwayapp/railpack/pull/289)

**Full Changelog**: [v0.8.0...v0.9.0](https://github.com/railwayapp/railpack/compare/v0.8.0...v0.9.0)

## v0.8.0
September 25, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.8.0)

### What's Changed
* chore: mise update 2025.9.17 by @iloveitaly in [#274](https://github.com/railwayapp/railpack/pull/274)
* feat: dockerignore support by @iloveitaly in [#263](https://github.com/railwayapp/railpack/pull/263)
* fix: missing local layer usage by @iloveitaly in [#276](https://github.com/railwayapp/railpack/pull/276)
* chore: mise update 2025.9.18 by @iloveitaly in [#275](https://github.com/railwayapp/railpack/pull/275)

**Full Changelog**: [v0.7.2...v0.8.0](https://github.com/railwayapp/railpack/compare/v0.7.2...v0.8.0)

## v0.7.2
September 19, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.7.2)

### What's Changed
* fix: gracefully handle resolution failures for packages with SkipMiseInstall=true by @coffee-cup in [#273](https://github.com/railwayapp/railpack/pull/273)

**Full Changelog**: [v0.7.1...v0.7.2](https://github.com/railwayapp/railpack/compare/v0.7.1...v0.7.2)

## v0.7.1
September 17, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.7.1)

### What's Changed
* Add configurable skippable commands list to pretty print output by @coffee-cup in [#271](https://github.com/railwayapp/railpack/pull/271)

**Full Changelog**: [v0.7.0...v0.7.1](https://github.com/railwayapp/railpack/compare/v0.7.0...v0.7.1)

## v0.7.0
September 16, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.7.0)

### What's Changed
* Update Rust examples by @coffee-cup in [#268](https://github.com/railwayapp/railpack/pull/268)
* Update default versions for Go, Python, Ruby, and Rust providers by @coffee-cup in [#267](https://github.com/railwayapp/railpack/pull/267)

**Full Changelog**: [v0.6.1...v0.7.0](https://github.com/railwayapp/railpack/compare/v0.6.1...v0.7.0)

## v0.6.1
September 16, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.6.1)

### What's Changed
* test: use poetry cli to run python by @iloveitaly in [#257](https://github.com/railwayapp/railpack/pull/257)
* chore: mise update 2025.9.9 by @iloveitaly in [#262](https://github.com/railwayapp/railpack/pull/262)
* fix: remove python3-dev apt package by @iloveitaly in [#256](https://github.com/railwayapp/railpack/pull/256)
* Update node examples by @railway-bot in [#266](https://github.com/railwayapp/railpack/pull/266)

### New Contributors
* @railway-bot made their first contribution in [#266](https://github.com/railwayapp/railpack/pull/266)

**Full Changelog**: [v0.6.0...v0.6.1](https://github.com/railwayapp/railpack/compare/v0.6.0...v0.6.1)

## v0.6.0
September 11, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.6.0)

### What's Changed
* feat: support for expanded platform arguments by @iloveitaly in [#246](https://github.com/railwayapp/railpack/pull/246)
* chore: mise update 2025.9.6 by @iloveitaly in [#254](https://github.com/railwayapp/railpack/pull/254)
* fix: hard fail on invalid railpack.json by @iloveitaly in [#226](https://github.com/railwayapp/railpack/pull/226)
* feat: add support for .bun-version file by @coffee-cup in [#258](https://github.com/railwayapp/railpack/pull/258)

**Full Changelog**: [v0.5.1...v0.6.0](https://github.com/railwayapp/railpack/compare/v0.5.1...v0.6.0)

## v0.5.1
September 5, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.5.1)

### What's Changed
* Revert "fix: use default mise backends for python packages" by @coffee-cup in [#252](https://github.com/railwayapp/railpack/pull/252)

**Full Changelog**: [v0.5.0...v0.5.1](https://github.com/railwayapp/railpack/compare/v0.5.0...v0.5.1)

## v0.5.0
September 4, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.5.0)

### What's Changed
* chore: mise update 2025.8.21 by @iloveitaly in [#236](https://github.com/railwayapp/railpack/pull/236)
* test: fix php snapshot by @iloveitaly in [#239](https://github.com/railwayapp/railpack/pull/239)
* fix(install): handle GitHub API throttling by @lionello in [#235](https://github.com/railwayapp/railpack/pull/235)
* fix: add -e alias for --env by @iloveitaly in [#243](https://github.com/railwayapp/railpack/pull/243)
* test: assert against exact yarn package manager versions by @iloveitaly in [#238](https://github.com/railwayapp/railpack/pull/238)
* fix: fail early if no shell start script is detected, stop globbing shell script by @iloveitaly in [#240](https://github.com/railwayapp/railpack/pull/240)
* fix: fallback to raw version when no package versions exist by @iloveitaly in [#245](https://github.com/railwayapp/railpack/pull/245)
* test: pass GITHUB_TOKEN to docker containers running the railpack buildplan by @iloveitaly in [#247](https://github.com/railwayapp/railpack/pull/247)
* fix: support node projects without dependencies by @iloveitaly in [#242](https://github.com/railwayapp/railpack/pull/242)
* feat: support react router build cache by @iloveitaly in [#244](https://github.com/railwayapp/railpack/pull/244)
* fix: improved buildkit error message by @iloveitaly in [#237](https://github.com/railwayapp/railpack/pull/237)
* refactor: use consistent local layers entrypoint by @iloveitaly in [#241](https://github.com/railwayapp/railpack/pull/241)
* fix: use default mise backends for python packages by @iloveitaly in [#249](https://github.com/railwayapp/railpack/pull/249)
* fix: --verbose enables mise verbose logging by @iloveitaly in [#248](https://github.com/railwayapp/railpack/pull/248)

### New Contributors
* @lionello made their first contribution in [#235](https://github.com/railwayapp/railpack/pull/235)

**Full Changelog**: [v0.4.0...v0.5.0](https://github.com/railwayapp/railpack/compare/v0.4.0...v0.5.0)

## v0.4.0
August 27, 2025 · [GitHub release](https://github.com/railwayapp/railpack/releases/tag/v0.4.0)

### What's Changed
* fix: remove yarn gpg workaround by @iloveitaly in [#220](https://github.com/railwayapp/railpack/pull/220)
* chore: mise update 2025.8.13 by @iloveitaly in [#224](https://github.com/railwayapp/railpack/pull/224)
* build: use railpack as the binary output name by @iloveitaly in [#227](https://github.com/railwayapp/railpack/pull/227)
* feat: add hidden dump-llb flag by @iloveitaly in [#225](https://github.com/railwayapp/railpack/pull/225)
* consistent provider step naming by @iloveitaly in [#228](https://github.com/railwayapp/railpack/pull/228)
* chore: mise update 2025.8.18 by @iloveitaly in [#229](https://github.com/railwayapp/railpack/pull/229)
* test: update snapshots for new caddy version by @iloveitaly in [#230](https://github.com/railwayapp/railpack/pull/230)
* fix: Support colocated hooks in Phoenix (LiveView) for Elixir by @jesse-c in [#231](https://github.com/railwayapp/railpack/pull/231)

### New Contributors
* @jesse-c made their first contribution in [#231](https://github.com/railwayapp/railpack/pull/231)

**Full Changelog**: [v0.3.0...v0.4.0](https://github.com/railwayapp/railpack/compare/v0.3.0...v0.4.0)

## Older releases

This page covers the last year. See all releases on [GitHub](https://github.com/railwayapp/railpack/releases).

