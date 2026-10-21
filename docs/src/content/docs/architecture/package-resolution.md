---
title: Packages and Version Resolution
description: Understanding how Railpack resolves package versions using Mise
---

Railpack providers will analyze the app and determine _fuzzy_ versions of
executables to install. Versions like `3.13`, or `22`. The version resolution
step will resolve those fuzzy versions into the latest valid version that
exists.

[Mise](https://mise.jdx.dev/) is used for the package resolution using the `mise
latest package@version` command. Mise is also used for (most) package
installations in the builds as well. However, this is not a requirement of
Railpack and alternative installation methods are possible (for example PHP will
use Mise to resolve a valid version and then start from a PHP base image).
Railpack enables Mise paranoid mode for stricter security validation.

For more information on how Railpack utilizes Mise and our philosophy on tool
defaults, see the [Mise Configuration](/config/mise) guide.

## Semver constraints from app manifests

Manifest constraints (for example `engines.pnpm`) are not resolved with
npm-style semver. Caret (`^`) and range (`>=`, `<`) notation are simplified to
the major version only (`^10.34.0` → `10`) before Mise picks a release. This
is done because Mise does not support compound version constraints right now.
Use an exact version or `mise.toml` to pin. See also
[`minimum_release_age`](/config/mise#minimum_release_age).

Platforms can preserve previously resolved defaults between builds. See
[Package Version Resolution](/platforms/package-version-resolution) for the
platform integration workflow.
