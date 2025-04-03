---
title: Secrets and Environment Variables
description: How Railpack handles secrets and environment variables
---

Build secrets and environment variables are treated separately. The main
differences being:

- Environment variables are saved in the final image and should not contain
  sensitive information. Since they are in the final image, providers can add
  variables that will be available to the app at runtime.
- Secrets are never logged or saved in the build logs. They are also only
  available at build time and not saved to the final image.

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

## Secrets

The names of all secrets that should be used during the build are added to the
top of the build plan, and each step's `secrets` array specifies which secrets
should invalidate that step's layer cache when their values change. While all
secrets are available to every command as environment variables, only the ones
listed in a step's `secrets` array will trigger a cache invalidation if
modified.

Under the hood, Railpack uses [BuildKit secrets
mounts](https://docs.docker.com/build/building/secrets/) to supply an exec
command with the secret value as an environment variable.

By default, all secrets defined in the build plan are available to each step.
You can explicitly specify which secrets a step should have access to using the
`secrets` array. An empty array indicates that no secrets should be available to
that step.

```json
{
  "secrets": ["DATABASE_URL", "API_KEY", "STRIPE_LIVE_KEY"],
  "steps": {
    "build": {
      "secrets": ["DATABASE_URL", "API_KEY"] // Only these secrets are available to this step
    }
  }
}
```

You can also use `"*"` in a step's secrets array to indicate that it should have
access to all secrets defined in the build plan:

```json
{
  "secrets": ["DATABASE_URL", "API_KEY", "STRIPE_LIVE_KEY"],
  "steps": {
    "build": {
      "secrets": ["*"] // This step has access to all secrets
    }
  }
}
```

### Providing Secrets

You can add secrets when building or generating a build plan with the `--env`
flag. The names of these variables will be added to the build plan as secrets.

#### CLI Build

If building with [the CLI](/guides/building-with-cli), Railpack will check that
all the secrets defined in the build plan have variables.

```bash
railpack build --env STRIPE_LIVE_KEY=sk_live_asdf
```

#### Custom Frontend

If building with a [custom frontend](/guides/building-with-custom-frontends),
you should still provide the secrets when generating the plan with `--env`. This
adds the secrets to the build plan. You then need to pass the secrets to Docker
or BuildKit with the `--secret` flag.

```bash
# Generate a build plan
railpack plan --env STRIPE_LIVE_KEY=sk_live_asdf --out test/railpack-plan.json

# Build with the custom frontend
STRIPE_LIVE_KEY=asdf123456789 docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f test/railpack-plan.json \
  --secret id=STRIPE_LIVE_KEY,env=STRIPE_LIVE_KEY \
  --build-arg secrets-hash=asdfasdf \
  examples/node-bun
```

For more information about running Railpack in production, see the [Running
Railpack in Production](/guides/running-railpack-in-production) guide.

### Layer Invalidation

By default, BuildKit will not invalidate a layer if a secret is changed. To get
around this, Railpack uses a hash of the secret values and mounts this as a file
in the layer. This will bust the layer cache if the secret is changed. Pass the
secret hash to BuildKit with the `--build-arg secrets-hash=<hash>` flag.
