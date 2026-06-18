---
title: Recommendations
description: Environment and tooling recommendations for Railpack-built images
---

## Timezone

Railpack does not set a timezone. Set `TZ` explicitly for your application.
`TZ=UTC` is a good default.

```json
{
  "deploy": {
    "variables": {
      "TZ": "UTC"
    }
  }
}
```

## Python with pipx

If you install `pipx` without also installing Python via mise, pipx will use
the system Python and pip from the Debian runtime image. That Python is old and
can cause compatibility issues with mise and other tooling.

Install Python via mise whenever you use pipx:

```json
{
  "packages": {
    "python": "latest",
    "pipx:httpie": "3.2.4"
  }
}
```

Or, even better, configure this in a `mise.toml` instead of a `railpack.json`

## Locale

Set `LANG=en_US.UTF-8` so applications and shell tools handle Unicode
correctly. Railpack does not bundle additional locales — only `en_US.UTF-8`
is available in the runtime image. If your application requires a different
locale, install the corresponding locale packages via
[`deploy.aptPackages`](/config/file).

```json
{
  "deploy": {
    "variables": {
      "LANG": "en_US.UTF-8"
    }
  }
}
```

## Use Mise for Language and Package Manager Versioning

Although mise supports extracting versions from language-specific configuration
files (`.node-version`, `.python-version`, `runtime.txt`, etc.), mise offers a
much more unified and streamlined way of managing these versions. We highly
recommend moving development configuration as much as you can to mise.

Specify versions of languages, package managers, tools, and other dependencies
in your mise config rather than relying on defaults or `latest`. Pinning
versions keeps local development, CI, and production aligned.

A `mise.toml` in your project keeps tool versions in one place:

```toml
[tools]
node = "22.14.0"
python = "3.13.2"
poetry = "2.1.1"
jq = "1.7.1"
```

See [Mise Configuration](/config/mise) for how Railpack detects and applies
mise config during builds.

## Prefer Mise Over Apt, pipx, and Other Installers

When you need a CLI tool or utility, prefer installing it with mise over Apt,
pipx, npm, or other package managers. Mise does not support every tool, but it
supports most external tools you are likely to need.

Using mise keeps tool versions consistent with the rest of your config and
avoids coupling your build to distro packages or a separate install path.

```toml
[tools]
jq = "1.7.1"
ripgrep = "14.1.1"
```

Reach for Apt when you need system libraries or packages that mise does not
provide (for example `libpq-dev` or `ffmpeg`). See [Installing Additional
Packages](/guides/installing-packages) for how to add mise and Apt packages to
a build.

## Use Mise Lockfiles

Add `locked = true` to your mise config. This ensures that the exact same
version is used for dev, CI, and production.

```toml
[tools]
node = "22.14.0"
python = "3.13.2"

[settings]
locked = true
```

Commit the generated `mise.lock` file alongside your `mise.toml`. Railpack
automatically includes `mise.lock` files in the build when present.
