---
title: Secrets and Environment Variables
description: How Railpack handles secrets and environment variables
---

Build secrets and environment variables are treated separately. The main
differences being:

- Environment variables are saved in the final image and should not contain
  sensitive information. Since they are in the final image, providers can add
  variables that will be available to the app at runtime.
- Secrets are environment variables which never logged or saved in the build logs. They are also *only*
  available at build time and not saved to the final image (and therefore not available at runtime).

Some examples of where you might use each:

- **Environment Variables**: `NODE_ENV`, `PYTHON_ENV`, `TZ`, etc.
- **Secrets**: `SENTRY_AUTH_TOKEN` (used to report a new build to Sentry, but not required at runtime). Or an API key used to collect static assets during the build. *Do not* include `DATABASE_URL`, `STRIPE_API_KEY`, or other secrets required at runtime here (unless you need them at build time).

## Environment Variables

Environment variables can be set in a couple ways:

1. Through step variables. In this case, the variable is available only during that
   step:

```json
{
  "steps": {
    "install": {
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

2. Through the deploy section for runtime variables. These variables are only available at runtime and will not be set during the build:

```json
{
  "deploy": {
    "env": {
      "NODE_ENV": "production"
    }
  }
}
```

3. Through the top-level `variables` field. These variables are available during all steps of
   the build *and* at runtime:

```json
{
  "env": {
    "TZ": "America/Los_Angeles"
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

You can add secrets when building or generating a build plan with the `--secret`
flag. The names of these variables will be added to the build plan as secrets.

#### CLI Build

If building with [the CLI](/guides/building-with-cli), Railpack will check that
all the secrets defined in the build plan have variables.

```bash
railpack build --secret SENTRY_AUTH_TOKEN=sk_live_asdf
```

You can also provide environment variables with the `--env` flag. These will available to all steps and in the final image (using `railpack.json` if you need to customize when a environment variable is available).

```bash
railpack build --env TZ=UTC --secret SENTRY_AUTH_TOKEN=sk_live_asdf
```

#### Custom Frontend

If building with a [custom frontend](/guides/building-with-custom-frontends),
you should still provide the secrets when generating the plan with `--env`. This
adds the secrets to the build plan. You then need to pass the secrets to Docker
or BuildKit with the `--secret` flag.

```bash
# Generate a build plan
railpack plan --secret STRIPE_LIVE_KEY=sk_live_asdf --out railpack-plan.json

# Build with the custom frontend
SENTRY_AUTH_TOKEN=asdf123456789 docker build \
  --build-arg BUILDKIT_SYNTAX="ghcr.io/railwayapp/railpack:railpack-frontend" \
  -f railpack-plan.json \
  --secret id=SENTRY_AUTH_TOKEN,env=SENTRY_AUTH_TOKEN \
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
