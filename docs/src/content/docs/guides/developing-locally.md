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

This command will also start a buildkit container (check out `mise.toml` in the root directory for more information).

Use the `cli` task to run the railpack CLI (this is like `railpack --help`)

```bash
mise run cli --help
```

If you want to compile a development build of railpack to use elsewhere on your machine:

```bash
mise run build

# add the railpack repo `bin/` directory to your path to use the newly-compiled railpack on your machine
export PATH="$PWD/bin:$PATH"
```

## Building directly with Buildkit

**👋 Requirement**: an instance of Buildkit must be running locally.
Instructions in "[Run BuildKit Locally](#run-buildkit-locally)" at the bottom of
the readme.

Railpack will instantiate a BuildKit client and communicate to over GRPC in
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

## Custom frontend

You can build with a [custom BuildKit frontend](/guides/custom-frontend), but
this is a bit tedious for local iteration.

The frontend needs to be built into an image and accessible to the BuildKit
instance. To see how you can build and push an image, see the
`build-and-push-frontend` mise task in `mise.toml`.

Once you have an image, you can do:

Generate a build plan for an app:

```bash
mise run cli plan examples/node-bun --out test/railpack-plan.json
```

Build the app with Docker:

```bash
docker buildx \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f test/railpack-plan.json \
  examples/node-bun
```

or use BuildKit directly:

```bash
buildctl build \
  --local context=examples/node-bun \
  --local dockerfile=test \
  --frontend=gateway.v0 \
  --opt source=ghcr.io/railwayapp/railpack:railpack-frontend \
  --output type=docker,name=test | docker load
```

_Note the `docker load` here to load the image into Docker. However, you can
change the [output](https://github.com/moby/buildkit?tab=readme-ov-file#output)
or push to a registry instead._

## Integration Tests

Integration tests build and run example applications in containers to verify
end-to-end functionality. Each example with a `test.json` file gets tested
automatically.

```bash
# Run all integration tests, this takes a long time. Let CI do this for you.
mise run test-integration

# Run specific test
mise run test-integration -- -run "TestExamplesIntegration/python-uv-tool-versions"
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

Integation tests can define services (postgres, redis, anything with a docker image) that
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
  just for railpack, and the mise binary run during the build process. The mise version will be the same, but the environment
  is different.
* If `mise tool erlang` reports a `core:` plugin it means this plugin is compiled into the mise binary and it's source is available with the mise monorepo. This can be confusing since there are often open source shell-based repos available for a tool as well, but they are unused by default.

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
```

## Debugging

Here's some helpful debugging tricks:

* `URFAVE_CLI_TRACING=on` for debugging CLI argument parsing
* `mise run cli -- --verbose build --show-plan --progress plain examples/node-bun`
* `mise run build`, add `./bin/` to your `$PATH`, and then run `railpack` in a separate local directory
* `docker exec buildkit buildctl prune` to clean the builder cache
* `NO_COLOR=1`

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

