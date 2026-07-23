---
title: Installing Additional Packages
description: Learn how to install additional packages in your build
---

Railpack supports installing additional versioned packages from
[Mise](https://mise.jdx.dev/), or packages from Apt.

## Mise

You can set the `RAILPACK_PACKAGES` environment variable to install additional
packages from Mise.

For example, this installs the latest versions of Node and Bun, and Python 3.10.

```bash
RAILPACK_PACKAGES="node bun@latest python@3.10"
```

You can find a list of available packages in the [Mise registry](https://mise.jdx.dev/registry.html).

## Apt

Apt packages are split into those needed for the build and those needed at
runtime.

You can set the `RAILPACK_BUILD_APT_PACKAGES` and
`RAILPACK_DEPLOY_APT_PACKAGES` environment variables to customize the Apt
packages installed during the build and deployment steps respectively.

In this example, we install `build-essential` during the build step and `ffmpeg`
at runtime while retaining packages Railpack adds automatically.

```bash
RAILPACK_BUILD_APT_PACKAGES="... build-essential"
RAILPACK_DEPLOY_APT_PACKAGES="... ffmpeg"
```

The `...` entry extends Railpack's generated package list. Without it, the
configured list replaces the packages Railpack specifies:

```json
{
  "$schema": "https://schema.railpack.com",
  "buildAptPackages": ["...", "build-essential"],
  "deploy": {
    "aptPackages": ["...", "ffmpeg"]
  }
}
```

To intentionally replace Railpack's runtime packages, omit `...`:

```json
{
  "$schema": "https://schema.railpack.com",
  "deploy": {
    "aptPackages": ["ffmpeg"]
  }
}
```
