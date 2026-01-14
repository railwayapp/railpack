---
title: Packages and Version Resolution
description: Understanding how Railpack resolves package versions using Mise
---

Railpack providers will analyze the app and determine _fuzzy_ versions of
executables to install. Versions like `3.13`, or `22`. The version resolution
step will resolve those fuzzy versions into the latest valid version that exists.

[Mise](https://mise.jdx.dev/) is used for the package resolution using the `mise
latest package@version` command. Mise is also used for (most) package
installations in the builds as well. However, this is not a requirement of
Railpack and alternative installation methods are possible (for example PHP will
use Mise to resolve a valid version and then start from a PHP base image).
Railpack enables Mise paranoid mode for stricter security validation.

## Previous and default versions

One important aspect of Railpack is that updating the default version of
executables in providers (e.g. Node 20 -> 22) should not change the version
installed for apps that have already been building successfully with
Railpack. This is mainly useful on platforms that use Railpack to build user
applications (e.g. Railway).

To support this, you can pass in a `--previous pkg@name ...` flag when
generating the build plan. The typical flow will go like this

- User builds for the first time with Railpack. The default Node version is used (20).
- The platform saves the resolved versions of packages used
- Railpack updates the default version of Node to 22
- User submits a new build. The platform passes a `--previous` flag
- Railpack will use the previous versions instead of using the new default version

This means that Railpack can freely update default versions of packages without
having to worry about breaking existing apps that rely on the previous defaults.

Passing in a previous version will only be used in place of the default. If a
more specific version of a package is requested (e.g. through a package.json
engines field or env var), then we will always use that.
