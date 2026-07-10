---
title: Developing Locally
description: Learn how to develop Railpack locally
---

Once you've [checked out the repo](https://github.com/railwayapp/railpack), you
can follow this to start developing locally.

## Getting Setup

We use [Mise](https://mise.jdx.dev/) for managing language dependencies and
tasks for building and testing Railpack. You don't have to use Mise, but it's
recommended.

Install and use all versions of tools needed for Railpack

```bash
# Assuming you are cd'd into the repo root
mise run setup
```

This command will also start a BuildKit container (check out `mise.toml` in the root directory for more information).

Use the `cli` task to run the Railpack CLI (this is like `railpack --help`)

```bash
mise run cli --help
```

If you want to compile a development build of Railpack to use elsewhere on your machine:

```bash
mise run build

# add the Railpack repo `bin/` directory to your path to use the newly-compiled Railpack on your machine
export PATH="$PWD/bin:$PATH"
```

## Building directly with BuildKit

**👋 Requirement**: an instance of BuildKit must be running locally.
Run `mise run setup` to start a BuildKit container.

Railpack will instantiate a BuildKit client and communicate over GRPC in
order to build the generated LLB.

```bash
mise run cli --verbose build examples/node-bun
```

Remember, `mise run` runs the cli in the root project directory. So, if you are in a specific project example directory, you'll want to specify the path to the example directory as an absolute path:

```bash
cd examples/node-angular/
mise run cli build $(pwd)
```

You need to have a BuildKit instance running (see below).

## Docker Images

Multiple Docker images are used in Railpack:

* **`images/alpine/frontend/`** for the Railpack BuildKit frontend. These are simple: they include a railpack binary in a image that can be executed by the buildpack frontend. One is designed to be built for production, one is for local testing and development. These are not used by the user's application during build or runtime.
* **`images/debian/*`:** for the Railpack build process. These are used within the buildpack exection of the railpack-generated llb.
  * `images/debian/build` used during the llb build process. These contain common tools, languages, mise, etc that might be used during the build process. Note that all of these utilities are *not* included in the final image in order to reduce the total image size.
  * `images/debian/runtime` a bare bones debian image used at runtime. The tools, build artifacts, etc generated during the railpack build are added to this base image.

## Custom frontend

You can build with a [custom BuildKit frontend](/guides/custom-frontend), but
this is a bit tedious for local iteration.

The frontend needs to be built into an image and accessible to the BuildKit instance:

```bash
mise run image-runtime-build
```

Then, generate a build plan for an app:

```bash
mise run cli plan examples/node-bun --out test/railpack-plan.json
```

With the image you built previously, you can now run the build:

```bash
docker buildx \
  --build-arg BUILDKIT_SYNTAX="railpack-frontend:local" \
  -f test/railpack-plan.json \
  examples/node-bun
```

You can also use the `buildctl` command to run BuildKit directly. This is helpful as it's a lower level command which
exposes helpful debugging flags. However, you can't reference the locally built image without loading it into a registry first.

This is done automatically for you with `image-runtime-build` but you must start the registry first with `image-run-registry`.

Then, you can run the build with the locally-build frontend:

```bash
buildctl build \
  --frontend=gateway.v0 \
  --opt source=railpack-frontend:local \
  --local context=examples/node-bun \
  --local dockerfile=test \
  --output type=docker,name=test | docker load
```

The `dockerfile=` param instructs railpack to use that directory to look for the `railpack-plan.json` file. The `context=` param is the path to the app to build. More specifically, `--local` 'uploads' the referenced directories
to the buildkit daemon.

Note the `docker load` here to load the image into Docker. However, you can
change the [output](https://github.com/moby/buildkit?tab=readme-ov-file#output)
or push to a registry instead.

You can also provide additional configuration to buildctl, like registry
cache import/export (use top-level flags, not `--opt`):

```bash
buildctl build \
  --frontend=gateway.v0 \
  --opt source=host.docker.internal:7890/railpack-frontend:local \
  --local context=examples/node-bun \
  --local dockerfile=tmp/frontend-plan \
  --export-cache type=registry,ref=host.docker.internal:7890/node-bun:cache,mode=max \
  --import-cache type=registry,ref=host.docker.internal:7890/node-bun:cache
```

Note that the cache arguments are different than what `docker buildx`. The equivalent `docker buildx` command would be:


```bash
docker buildx build \
  --build-arg BUILDKIT_SYNTAX="host.docker.internal:7890/railpack-frontend:local" \
  --cache-to=type=registry,ref=host.docker.internal:7890/node-bun:cache,mode=max \
  --cache-from=type=registry,ref=host.docker.internal:7890/node-bun:cache \ 
  -f tmp/frontend-plan/railpack-plan.json \
  examples/node-bun
```

Debugging a buildkit related problem? Enable debug logging:

```bash
buildctl --debug build \
  --frontend=gateway.v0 \
  --opt source=railpack-frontend:local \
  --local context=examples/node-bun \
  --progress=plain \
  --trace=tmp/builtctl-build-trace.log \
  --debug-json-cache-metrics stdout
```

Quick note about `builtctl` vs `docker buildx`. These two ways of invoking the railpack frontend handle arguments differently:

* `--build-arg` prefixes the argument with `build-arg:`.
* `--opt` does not prefix the build arg at all. You must prefix args with `build-arg:` if they are
  arguments handled by the railpack frontend.

## Integration Tests

Integration tests build and run example applications in containers to verify
end-to-end functionality. Each example with a `test.json` file gets tested
automatically.

```bash
# Run all integration tests, this takes a long time. Let CI do this for you.
mise run test-integration

# Run specific test
mise run test-integration -- -run "TestExamplesIntegration/python-uv-tool-versions"

# Or, from within an examples/ directory, run the test for that example
cd examples/python-uv-tool-versions
mise run test-integration-cwd
```

The `test.json` file contains an array of test cases. Each case builds and runs the same
image but checks for different expected output strings.

### HTTP Checks

In addition to a basic `justBuild: true` check or an output assertion, you can also run an HTTP check that starts the container and asserts that a specific route returns an expected HTTP code:

```json
{
  "httpCheck": {
    "path": "/",
    "expected": 200,
    "internalPort": 3000
  }
}
```

### Output Assertions

You can verify that the application outputs specific strings. `expectedOutput` can
be a single string or an array of strings that all must be present in the output:

```json
{
  "expectedOutput": "Server running on port 3000"
}
```

Or with multiple strings:

```json
{
  "expectedOutput": [
    "Elixir version: 1.18",
    "Erlang/OTP version: 27"
  ]
}
```

### Environment Variables

You can pass environment variables to the container at runtime using the
`envs` key. This is useful for testing with different configurations, secrets,
or Railpack configuration variables:

```json
{
  "expectedOutput": "Server running on port 3000",
  "envs": {
    "DATABASE_URL": "postgresql://user:password@postgres:5432/db",
    "SECRET_KEY": "test-secret"
  }
}
```

You can also use `RAILPACK_*` configuration variables in `envs` to test
different build configurations:

```json
{
  "expectedOutput": "hello from Node",
  "envs": {
    "RAILPACK_PRUNE_DEPS": "true",
    "RAILPACK_STATIC_FILE_ROOT": "/custom/path"
  }
}
```

See the [environment variables
documentation](/config/environment-variables) for a complete list of available
`RAILPACK_*` configuration options.

### Services

Integration tests can define services (postgres, redis, anything with a docker image) that
are required for the application to run. Create a `docker-compose.yml` in a test directory
and it will automatically be picked up and run before the project container is run.

Here's an example of how to run the container locally to manually test it:

```shell
docker compose up -d
docker run -it --network python-django_default --env DATABASE_URL="postgresql://django_user:django_password@postgres:5432/django_db" python-django
```

## Mise

Mise is absolutely central to this entire project, so you'll have to dig into the details.

* `mise trust` state is located in `~/.local/state/mise/trusted-configs`
* There are two mise 'environments' to keep in mind: the host environment, which uses a specific version of mise downloaded
  just for Railpack, and the mise binary run during the build process. The mise version will be the same, but the environment
  is different.
* If `mise tool erlang` reports a `core:` plugin it means this plugin is compiled into the mise binary and its source is available with the mise monorepo. This can be confusing since there are often open source shell-based repos available for a tool as well, but they are unused by default.

### Mise Commands

Some helpful commands for debugging issues with mise:

```bash
# Lint and format
mise run check

# Where is a particular binary?
mise where pipx:squawk-cli@

# Run tests
mise run test

# Start the docs dev server
mise run docs-dev

# Inspect what backend is being used for a given tool
mise tool poetry

# test a tool out without adding it to your environment
mise exec pipx:httpie -- http google.com
```

## Debugging

Here's some helpful debugging tricks:

* `URFAVE_CLI_TRACING=on` for debugging CLI argument parsing
* `RAILPACK_DEBUG=1` for debugging Railpack debug logging
* `--build-arg verbose=true` for debugging the frontend (or `--opt build-arg:verbose=true` with `buildctl`)
* `docker logs -f buildkit` to see the BuildKit daemon logs, which includes railpack logs when it's used as a frontend
* `docker logs -f railpack-registry` to inspect local registry logs. Helpful for debugging cache import/export issues.
* `mise run cli -- --verbose build --show-plan --progress plain examples/node-bun`
* `mise run build`, add `./bin/` to your `$PATH`, and then run `railpack` in a separate local directory
* `docker exec buildkit buildctl prune` to clean the builder cache
* `NO_COLOR=1`

### Inspecting LLB Output

The `--dump-llb` flag outputs the raw BuildKit LLB (Low-Level Builder)
definition, which can be piped to various tools for inspection:

#### Visualize LLB as a graph

```bash
mise run cli build $(pwd) --dump-llb | \
  buildctl debug dump-llb --dot | \
  dot -Tpng > graph.png
```

#### Inspect LLB as JSON

```bash
mise run cli build $(pwd) --dump-llb | \
  buildctl debug dump-llb | \
  fx
```

_Note: Any JSON visualization tool can be used (jq, fx, jless, etc.)_

#### Build directly with buildctl

```bash
mise run cli build $(pwd) --dump-llb | \
  buildctl build \
    --progress=plain \
    --trace=build.log \
    --local context=.
```

### Interactive Debugging with Delve

```sh
mise run debug-cli build $(pwd)
```

Then, set some breakpoints:

```
break core/providers/node/node.go:177
continue
```

The commands you probably want: `ls`, `print build.Commands`, `continue`, `next`, `locals`,

## Node

The node provider is the most complex.

### Corepack

* `corepack` used to be included by default in node. It was removed in node 25. Now it must be installed via `npm install -g`
* corepack does not support node 25. >= 26 is officially required.
* corepack is installed into `node_modules` but the package managers that corepack installs are added to `COREPACK_HOME`
  which we customize to be a `/opt` path.
* `corepack prepare` generates shims next to the `node` binary. These symlink to `.js` files.
* We detect corepack usage based on the `package.json > engines > pnpm, etc` fields. If we find this, we `npm install -g` corepack for the user. However, this happens *after* the mise install step and the corepack commands end up mutating the global node_modules dirs.
* In order to make sure these changes are picked up by future steps (i.e. if `pnpm run` is used in a `startCommand`) we have to include the mise shims folder and the mise node folder from the build step.
  * All shims are installed into `/mise/shims`. There are no subfolders.

## Maintenance

There are some manual maintenance tasks that need to be done periodically:

* Mise versions need to updated
* Test snapshots which use `latest` for runtime versions need to be updated periodically.
* Elixir<>OTP version map needs to be updated as new major versions come out.
* PNPM lockfile versions are manually mapped to minimum pnpm versions
* Pnpm default version needs to be updated as LTS versions are released.
* Node default version needs to be updated as LTS versions are released.
