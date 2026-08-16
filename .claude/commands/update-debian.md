Update Railpack's builder and runtime images to a new Debian stable release. Focus on code changes and tests. If `$ARGUMENTS` is a Debian codename (e.g. `forky`), use that; otherwise pick the current Debian stable.

Do not commit, push, or open a PR unless asked.

## 1. Pick the target

Confirm the current stable Debian version and codename (wiki.debian.org/DebianReleases). Note the **old** and **new** codenames from:

- `images/debian/build/Dockerfile` (`FROM buildpack-deps:<codename>-scm`)
- `images/debian/runtime/Dockerfile` (`FROM debian:<codename>-slim`)
- `.github/workflows/release.yml` (`debian-<codename>` tags)

## 2. Image and tag changes

Update all three to the new codename. Keep the `-scm` / `-slim` suffixes.

CI only rebuilds and pushes `ghcr.io/railwayapp/railpack-{builder,runtime}:mise-*` when `core/mise/version.txt` is **newer than origin/main**. Dockerfile `FROM` changes do nothing in production until that happens. **Never overwrite an existing `mise-*` tag.**

- If `version.txt` already differs from origin/main, the Debian change can ride that bump.
- Otherwise set `core/mise/version.txt` to the latest mise release (`gh release view --repo jdx/mise --json tagName --jq .tagName`, strip a leading `v`).

Leave the comment in `.github/workflows/integration-tests.yml` that explains this coupling.

## 3. Companion code

Search the repo for the old codename and for hardcoded apt names. Typical call sites:

| Area | Files |
| --- | --- |
| FrankenPHP image tag | `core/providers/php/php.go` (`dunglas/frankenphp:php%s-<codename>`) |
| Node Chromium deps | `core/providers/node/node.go` (puppeteer + playwright lists) |
| Python native / Playwright deps | `core/providers/python/python.go` |
| Other provider apt | `core/providers/{ruby,golang,dotnet,shell}` |
| Builder apt packages | `images/debian/build/Dockerfile` |
| Example apt packages | `examples/**/railpack*.json` |
| Docs | `docs/src/content/docs/architecture/overview.md` |

Do **not** change `core/generate/context_test.go` deploy-base fixtures unless they are asserting Railpack's default image; those tests use a user-configured `debian:*-slim`.

Do **not** edit the generated changelog.

### FrankenPHP

`getPhpImage` bakes the Debian codename into the tag. Confirm Docker Hub has `dunglas/frankenphp:php<version>-<new-codename>` for the PHP versions used in `examples/php-*` (the provider also probes Hub at plan time). Update the tag format **and** the nearby comments.

### Apt packages

Debian often removes packages, adds `t64` names, or replaces transitional packages that `apt-cache` still lists but `apt-get install` rejects.

Collect every package from the Dockerfiles, provider lists, and example `aptPackages` / `buildAptPackages`. Install them for real on both new bases (not `apt-cache policy`):

```bash
docker run --rm debian:<new>-slim bash -lc 'apt-get update && apt-get install -y --no-install-recommends <pkgs>'
docker run --rm buildpack-deps:<new>-scm bash -lc 'apt-get update && apt-get install -y --no-install-recommends <pkgs>'
```

Use names that `apt-get install` accepts on the **new** distro. If a package is gone (e.g. `neofetch` on Trixie), replace it in examples with something that still exists and update `test.json` expected output.

Do not add comments that explain what the previous Debian release dropped or renamed. Keep that for the PR description.

## 4. Local images for CLI / integration tests

`mise run cli` and integration tests pull `ghcr.io/railwayapp/railpack-{builder,runtime}:mise-$(cat core/mise/version.txt)`, not `:local`. Build, then retag so those refs resolve to the new Dockerfiles:

```bash
mise run image-builder-build
mise run image-runtime-build
ver=$(cat core/mise/version.txt)
docker tag railpack-builder:local "ghcr.io/railwayapp/railpack-builder:mise-${ver}"
docker tag railpack-runtime:local "ghcr.io/railwayapp/railpack-runtime:mise-${ver}"
```

Confirm `/etc/os-release` on the runtime image shows the new `VERSION_CODENAME`.

## 5. Tests

After code changes:

```bash
mise run check
mise run test-update-snapshots
# review snapshot diffs (FrankenPHP tags, apt install command strings)
mise run test
```

Then build **and run** examples that install apt packages or use FrankenPHP. Prefer `mise run test-integration -- -run 'TestExamplesIntegration/<example>'` (Playwright `test.json` envs are applied automatically). Minimum set:

- `shell-script` — new runtime base
- `config-file` — user apt packages
- `python-system-deps` — cairo / poppler / ffmpeg
- `rust-system-deps` — OpenSSL on the builder
- `node-lts-npm-native-deps` — native modules on the builder
- `node-puppeteer` (including the custom-apt `test.json` case)
- `node-playwright` / `python-playwright`
- `php-vanilla`, `php-vanilla-82`, `php-laravel-12-react`

Fix install failures (`Unable to locate package`, `has no installation candidate`) in code, not by skipping the example.

## 6. Report

Old → new Debian. Whether mise was bumped and to which version. Companion package / FrankenPHP / example changes. Which integration tests ran and what failed.

Include PR-description context for package churn (removed, renamed, or transitional), for example: Debian 13 dropped `gconf-service` / `libgconf-2-4` / `libappindicator1`, renamed `libgdk-pixbuf2.0-0` to `libgdk-pixbuf-2.0-0`, and removed `neofetch` from the archive. Do not put that prose in source comments.
