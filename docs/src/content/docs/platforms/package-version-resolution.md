---
title: Package Version Resolution
description: Preserve resolved package versions across builds on a platform
---

This page covers preserving resolved package versions when building applications
on a platform. For details about how application configuration is interpreted
and how versions are resolved, see [Packages and Version
Resolution](/architecture/package-resolution).

## Previous and Default Versions

Updating the default version of an executable in a provider (for example, Node
20 to Node 22) should not change the version installed for applications that
have already been building successfully.

To preserve the previously resolved defaults, pass one or more
`--previous package@version` flags when generating the build plan. A typical
build flow is:

- A user builds for the first time and the current default Node version is used.
- The platform saves the resolved package versions from the build information.
- A later release changes the default Node version.
- The user submits another build and the platform passes the saved versions with
  `--previous`.
- The saved versions are used instead of the new defaults.

This allows default package versions to change without breaking existing
applications that rely on the previous defaults.

A previous version is only used in place of a provider default. If the
application requests a more specific version through a manifest, configuration
file, or environment variable, that explicit request takes precedence.

Railway applies this workflow automatically.
