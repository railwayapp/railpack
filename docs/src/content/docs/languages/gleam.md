---
title: Gleam
description: Building Gleam applications with Railpack
---

Railpack builds and deploys Gleam applications with zero configuration.

## Detection

Your project will be detected as a Gleam application if a `gleam.toml` file exists in the root directory.

## Versions

Both Gleam and Erlang default to the latest version; Erlang is available in both the build and runtime environments, but Gleam is only available during the build. This can be overridden with any mise-supported version file (`mise.toml`, `.tool-versions`, etc) or the `RAILPACK_PACKAGES` environment variable.

## Configuration

Railpack will build your Gleam application as an Erlang shipment using `gleam export erlang-shipment`, and run it using `./build/erlang-shipment/entrypoint.sh run`. By default, the source tree is not available in the final container.

Variables available:
- `RAILPACK_GLEAM_INCLUDE_SOURCE` - if this variable is truthy, the source tree will be included in the final container.
