---
title: Secrets and Environment Variables
description: How Railpack handles secrets and environment variables
---

Build secrets and environment variables are treated separately. The main
differences being:

- Environment variables are saved in the final image and should not contain
  sensitive information. Since they are in the final image, providers can add
  variables that will be available to the app at runtime.
- Secrets are available at build time and are not automatically saved to the
  final image. Build commands can still print secrets or persist them in build
  artifacts, and secret values are not automatically filtered from logs, so you
  are responsible for preventing build commands and artifacts from exposing
  them.

## Environment Variables

Environment variables can be set in two ways:

1. Through step variables:

```json
{
  "steps": {
    "install": {
      "variables": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

2. Through the deploy section for runtime variables:

```json
{
  "deploy": {
    "variables": {
      "NODE_ENV": "production"
    }
  }
}
```

## Runtime Railpack Variables

The final image includes runtime metadata that is not available to build
commands:

| Name                | Description                              |
| :------------------ | :--------------------------------------- |
| `RAILPACK_VERSION`  | Version used to produce the image        |
| `RAILPACK_BUILT_AT` | Build time as Unix epoch seconds (UTC)   |

Runtime metadata is added after generating the build graph, so it does not
invalidate application layer caches. The `RAILPACK_` namespace in the final
image is reserved for generated runtime metadata, which takes precedence over
deploy variables with the same name.

## Secrets

The generated build plan has a top-level catalog containing the names of every
secret that can be used during the build. You can define names in the root
`secrets` array or directly in a step's `secrets` array. Explicit secrets on
retained steps are automatically promoted into the generated plan's catalog
and do not need to be repeated at the root.

Under the hood, Railpack uses [BuildKit secrets
mounts](https://docs.docker.com/build/building/secrets/) to supply an exec
command with the secret value as an environment variable.

The frontend mounts every secret in the plan catalog into every exec command.
A step's `secrets` array does not restrict access. It selects the values that
invalidate that step's layer cache. An empty array disables secret-based cache
invalidation for the step, but cataloged secrets remain available to its exec
commands.

```json
{
  "steps": {
    "build": {
      "secrets": ["DATABASE_URL", "API_KEY"]
    }
  }
}
```

Use `"*"` in a step's `secrets` array to invalidate it when any root-configured
secret changes. Secrets inferred from other steps are not included in this
selection:

```json
{
  "secrets": ["DATABASE_URL", "API_KEY", "STRIPE_LIVE_KEY"],
  "steps": {
    "build": {
      "secrets": ["*"]
    }
  }
}
```

### Providing Secrets

You can add secrets when building or generating a build plan with the `--env`
flag. The names of these variables will be added to the build plan as secrets.

#### CLI Build

If building with [the CLI](/reference/cli/#build), Railpack will check that all
the secrets defined in the build plan have variables.

```bash
export STRIPE_LIVE_KEY=sk_live_asdf
railpack build --env STRIPE_LIVE_KEY
```

#### Custom Frontend

If building with the [BuildKit frontend](/platforms/buildkit-frontend), secret
names can come from `railpack.json` or `--env` during plan generation. You must
then provide values for every name in the generated plan catalog to Docker or
BuildKit with `--secret`.

```bash
# Generate a build plan
export STRIPE_LIVE_KEY=sk_live_123
railpack plan --env STRIPE_LIVE_KEY --out test/railpack-plan.json

# Hash the same value that will be supplied to BuildKit
secrets_hash=$(
  echo -n "STRIPE_LIVE_KEY=sk_live_123" |
    sha256sum |
    awk '{print $1}'
)

# Build with the custom frontend
docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack-frontend" \
  -f test/railpack-plan.json \
  --secret id=STRIPE_LIVE_KEY,env=STRIPE_LIVE_KEY \
  --build-arg secrets-hash="$secrets_hash" \
  examples/node-bun
```

For more information about running Railpack in production, see the [Running
Railpack in Production](/platforms/running-railpack-in-production) guide.

### BuildKit Secret Filtering

BuildKit does not make every secret passed on the command line automatically
available to a frontend. The frontend requests values by ID using the names in
the plan catalog:

- A cataloged and supplied secret is mounted into every exec command.
- A supplied secret absent from the catalog is never requested or mounted.
- A cataloged secret that was not supplied causes the build to fail because
  Railpack requests required secret mounts.

BuildKit's frontend API does not provide a way to enumerate every supplied
secret ID. Railpack can request a known ID but cannot discover undeclared
secrets from the BuildKit session.

## Secrets & Layer Invalidation

BuildKit intentionally excludes secret contents from an exec operation's cache
key. Without additional input, changing a mounted secret can reuse a layer
created with its previous value.

Railpack accepts an opaque hash through the `secrets-hash` build argument and
uses it to create cache dependencies. The hash must be deterministic and must
change whenever any supplied value represented by it changes. The CLI computes
this automatically. Frontend users must pass it with
`--build-arg secrets-hash=<hash>` for Docker or
`--opt build-arg:secrets-hash=<hash>` for `buildctl`.

The compiled `secrets` list on each step determines how the hash is used:

- An empty list adds no secret-based cache dependency.
- Concrete names derive a hash from only those mounted values. Changes to other
  secrets do not invalidate the step's resulting layer.
- `"*"` depends directly on the complete supplied `secrets-hash`.

When compiling `railpack.json`, `"*"` normally expands to the concrete root
secret names. This keeps secrets inferred from another step from invalidating
the wildcard step. The selector is preserved only when the generated plan has
no concrete secret catalog. A manually authored `railpack-plan.json` may also
contain `"*"`; the frontend interprets it as the complete supplied hash.

The value passed as `secrets-hash` can include values that are not present in
the plan catalog. Those values are still not mounted. However, a preserved
`"*"` will invalidate when that complete hash changes, including changes caused
by such extra values.
