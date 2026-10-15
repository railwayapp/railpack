---
title: Zig
description: Building Zig applications with Railpack
---

Railpack builds and deploys Zig applications.

## Detection

Your project will be detected as a Zig application if a `build.zig` file exists
in the root directory.

## Versions

Zig defaults to the latest version. This can be overridden with any
mise-supported version file (`mise.toml`, `.tool-versions`, etc) or the
`RAILPACK_PACKAGES` environment variable.

## Configuration

Railpack builds your application with `zig build --release=safe`. Only
`zig-out/bin` is copied into the final container, so the source tree and the
build cache are not shipped.

The start command is `./zig-out/bin/<name>`. `<name>` is the `.name` field in
`build.zig.zon`, or the application directory name when that file is absent.
Override it with `RAILPACK_START_CMD`:

```sh
RAILPACK_START_CMD="./zig-out/bin/zig_example"
```
