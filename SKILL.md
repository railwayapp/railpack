---
name: railpack
description: >-
  Configure and troubleshoot Railpack builds, with emphasis on RAILPACK_*
  environment variables, railpack.json overlays, build-plan inspection, local
  CLI installation and usage, and local BuildKit containers. Use for Railpack
  provider configuration, custom install/build/start commands, Mise or Apt
  packages, build or runtime variables and secrets, generated-plan debugging,
  BUILDKIT_HOST errors, or running Railpack from a release or source checkout.
---

# Railpack

Customize the generated BuildKit LLB without introducing a Dockerfile. Do not
propose a Dockerfile as the solution to a Railpack configuration problem.

## Use current sources

Fetch <https://railpack.com/llms.txt> as the documentation index before relying
on Railpack behavior. Open only the pages relevant to the task. Search
<https://railpack.com/llms-full.txt> when the answer spans several sections.

Prioritize:

- Environment Variables plus the detected language's page for `RAILPACK_*`.
- Configuration File for `railpack.json` fields and merge behavior.
- Secrets and Environment Variables for build-time versus runtime values.
- CLI Reference, Installation, and Getting Started for local commands.

When working in a Railpack source checkout, also inspect its instructions,
`mise.toml`, current documentation source, and CLI implementation. The checkout
may be ahead of the published docs. Verify flags with `railpack --help` or the
checkout's CLI task. Link the relevant public docs in user-facing guidance.

## Follow the configuration workflow

1. Inspect the application manifests, lockfiles, version files, start scripts,
   existing `railpack.json`, and deployment configuration. Do not expose secret
   values while inspecting the environment.
2. Generate a baseline plan before changing configuration when the CLI is
   available:

   ```sh
   railpack info .
   railpack plan .
   ```

3. Use a documented `RAILPACK_*` variable for a simple supported override. Use
   `railpack.json` for structured changes, additions to generated arrays,
   custom steps, layers, caches, variables, or secrets.
4. Generate the plan again and compare the relevant steps, deploy inputs,
   packages, variables, and start command.
5. Run a BuildKit build only when execution is part of the request.

Prefer the smallest targeted override. Do not set the same option through CLI
flags, environment variables, and `railpack.json`; mixed configuration obscures
which value wins.

## Configure with `RAILPACK_*`

Confirm every name and value format in the shared environment-variable page or
the relevant language page. Common build-plan variables include:

- `RAILPACK_INSTALL_CMD`
- `RAILPACK_BUILD_CMD`
- `RAILPACK_START_CMD`
- `RAILPACK_PACKAGES`
- `RAILPACK_BUILD_APT_PACKAGES`
- `RAILPACK_DEPLOY_APT_PACKAGES`
- `RAILPACK_DISABLE_CACHES`
- `RAILPACK_CONFIG_FILE`

Treat this list as a routing aid, not a complete reference. Provider-specific
variables control language versions, package managers, framework selection, and
provider features.

Understand replacement semantics:

- `RAILPACK_INSTALL_CMD` and `RAILPACK_BUILD_CMD` replace provider-generated
  command lists; they do not append.
- `RAILPACK_START_CMD` replaces the detected start command.
- Use `railpack.json` with `"..."` when adding a command while retaining
  generated commands.

Use space-separated values for variables documented as lists. Include `...`
when adding Apt packages without replacing Railpack's generated packages:

```sh
RAILPACK_PACKAGES="node@22 jq@latest"
RAILPACK_BUILD_APT_PACKAGES="... build-essential libssl-dev"
RAILPACK_DEPLOY_APT_PACKAGES="... ffmpeg"
```

For local `plan`, `info`, and `build` commands, pass build-plan variables
through `--env`. The CLI does not automatically import arbitrary exported
variables:

```sh
railpack plan \
  --env 'RAILPACK_PACKAGES=node@22 jq@latest' \
  --env 'RAILPACK_START_CMD=node dist/server.js' \
  .

export RAILPACK_NODE_VERSION=22
railpack build --env RAILPACK_NODE_VERSION .
```

Prefer name-only `--env SECRET_NAME` after exporting a secret so its value is
not placed in the command arguments. Treat values supplied through `--env` as
build-time secrets, not runtime configuration.

Set non-secret build-step values in `steps.<name>.variables`. Set values that
must be baked into the final image in `deploy.variables`; prefer the hosting
platform's runtime environment for environment-specific values. Never put
secret values in `railpack.json`.

Treat `RAILPACK_VERBOSE` separately: it is a process-level CLI option. Use
`RAILPACK_VERBOSE=1 railpack build ...` or `railpack --verbose build ...`, not
`--env RAILPACK_VERBOSE=1`.

## Construct `railpack.json` as an overlay

Place `railpack.json` at the build-context root. For another project-relative
path, prefer `--config-file path/to/config.json`; when using
`RAILPACK_CONFIG_FILE` with the local CLI, pass it through `--env`.

Always include the schema:

```json
{
  "$schema": "https://schema.railpack.com"
}
```

Write the smallest overlay on the provider-generated plan. Do not copy the
entire generated plan into the config file. Use `"..."` at the desired position
in a configurable array to retain generated values; omitting it replaces that
array.

```json
{
  "$schema": "https://schema.railpack.com",
  "packages": {
    "node": "22"
  },
  "buildAptPackages": ["...", "build-essential"],
  "steps": {
    "build": {
      "commands": ["...", "npm run postbuild"]
    }
  },
  "deploy": {
    "aptPackages": ["...", "ffmpeg"],
    "startCommand": "node dist/server.js"
  }
}
```

Use the documented model:

- Set `provider` only to override failed or ambiguous autodetection.
- Set `packages` for Mise-managed tools and versions.
- Set root `buildAptPackages` for build-only Apt packages.
- Set `steps.<name>` for inputs, commands, caches, variables, assets, secrets,
  and deploy outputs.
- Set `deploy` for runtime base, inputs, Apt packages, paths, variables, and
  start command.
- Filter layer inputs with `include` and `exclude` when only selected artifacts
  should enter another step or the runtime image.

Validate names and shapes against <https://schema.railpack.com>. Use plan
generation as the semantic validation:

```sh
railpack plan --out tmp/railpack-plan.json .
railpack info --format json --out tmp/railpack-info.json .
```

Inspect the generated output for accidental replacement of commands, Apt
packages, inputs, paths, caches, or secrets. Keep generated plan files separate
from the source `railpack.json` overlay.

## Install and run locally

Check for an existing binary first:

```sh
command -v railpack
railpack --version
railpack --help
```

If installation is requested, choose a current documented method. With Mise:

```sh
mise use github:railwayapp/railpack@latest
railpack --help
```

Or use the official release installer:

```sh
curl -sSL https://railpack.com/install.sh | sh
railpack --help
```

Pin the Railpack version when reproducibility matters. Consult the Installation
page for current version and destination controls instead of inventing flags.

Inside the Railpack repository, follow its checked-in instructions and
`mise.toml`. Prefer its development task over an installed release:

```sh
mise install
mise run setup
mise run cli -- --help
```

Use `railpack plan` or `mise run cli -- plan` without BuildKit. Use
`railpack build` for a normal local image build; reserve `prepare` and the
BuildKit frontend workflow for platform integrations.

## Run a local BuildKit container

Require a working Docker engine with Linux containers. Starting a privileged
container changes local Docker state; do it only when a local build is within
scope.

Inspect any existing container named `buildkit` before reusing the name. Reuse
or start an appropriate existing daemon, or choose another name and use the same
name in `BUILDKIT_HOST`. Do not remove an unknown container to resolve a name
collision.

For a disposable local daemon:

```sh
docker run --rm --privileged --detach \
  --name buildkit \
  moby/buildkit:latest

docker exec buildkit buildctl debug workers
```

Pass `BUILDKIT_HOST` in the same shell invocation as the build. This also works
when command executions do not share exported shell state:

```sh
BUILDKIT_HOST='docker-container://buildkit' \
  railpack build \
    --name my-app \
    --show-plan \
    --progress plain \
    ./path/to/project
```

The URI's final component must match the container name. Pin the BuildKit image
for reproducible environments. If the source checkout provides
`mise run setup`, prefer that task because it may mount repository-specific
BuildKit configuration and set `BUILDKIT_HOST`.

Use `docker logs buildkit` for daemon or connection failures. Stop the container
only if it was created for the task; `docker stop buildkit` also removes this
disposable `--rm` container.

## Troubleshoot in order

1. Verify the installed or development CLI and relevant subcommand help.
2. Generate a plan without BuildKit and resolve detection or config errors.
3. Verify every `RAILPACK_*` variable, its list format, and whether it replaces
   generated behavior.
4. Verify every modified array retains `"..."` unless replacement is intended.
5. Confirm Docker is reachable, the BuildKit worker is ready, and
   `BUILDKIT_HOST` names the correct container.
6. Re-run with `railpack --verbose build`, `--show-plan`, and plain progress.
7. Inspect BuildKit logs and Docker credentials for private images.

Identify the failing layer before broadening configuration. Explain the cause
and keep the fix targeted.
