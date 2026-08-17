---
title: Zig
description: Building Zig applications with Railpack
---

Railpack builds and deploys Zig applications with zero configuration.

## Detection

Your project will be detected as a Zig application if a `build.zig` file exists
in the root directory.

## Versions

Zig defaults to the latest version. This can be overridden with any
mise-supported version file (`mise.toml`, `.tool-versions`, etc) or the
`RAILPACK_PACKAGES` environment variable.

## Configuration

Railpack builds your application with `zig build --release=safe` and runs the
executable that is installed into `zig-out/bin`. The executable is named after
the `.name` field in `build.zig.zon`, and falls back to the name of the
application directory when that file is absent.

Only `zig-out/bin` is copied into the final container, so the source tree and
the build cache are not shipped.

Variables available:

- `RAILPACK_ZIG_RELEASE_MODE` - the release mode passed to `zig build`
  (`safe`, `fast`, or `small`). Defaults to `safe`.
