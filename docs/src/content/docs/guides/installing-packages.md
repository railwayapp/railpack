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

### Packages from a Third-Party Apt Repository

If a package is available from an Apt repository that is not configured by
default, Mise bootstrap can add the repository before installing the package.
Use a `pre-packages` hook to configure the repository, then declare the package
in `bootstrap.packages` in your `mise.toml`:

```toml
[bootstrap.hooks.pre-packages]
run = """
set -eu

install -m 0755 -d /etc/apt/keyrings
curl --fail --location https://packages.example.com/key.gpg \
  | gpg --dearmor --yes --output /etc/apt/keyrings/example.gpg

echo "deb [signed-by=/etc/apt/keyrings/example.gpg] \
https://packages.example.com/debian stable main" \
  > /etc/apt/sources.list.d/example.list
"""

[bootstrap.packages]
"apt:example-package" = "latest"
```

Railpack automatically runs the package portion of Mise bootstrap before
installing language tools. Apt packages in `bootstrap.packages` are available
during the build and in the runtime image.

Package hooks run in the build image only. Railpack carries Apt repository
definitions from `/etc/apt/sources.list.d` and keys from `/etc/apt/keyrings` or
`/usr/share/keyrings` into the runtime image before applying the packages there.
This avoids adding repository setup tools such as `curl` and `gpg` to the
runtime image.

See Mise's
[`bootstrap.packages`](https://mise.jdx.dev/bootstrap/packages/apt.html)
documentation for the supported Apt package syntax. Mise's `bootstrap.repos`
configuration is for Git repositories, not Apt repositories.
