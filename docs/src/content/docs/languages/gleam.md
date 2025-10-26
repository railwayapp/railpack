---
title: Gleam
description: Building Gleam applications with Railpack
---

Railpack builds and deploys Gleam applications with zero configuration.

## Detection

Your project will be detected as a Gleam application if a `gleam.toml` file exists in the root directory.

## Versions

Both Gleam and Erlang default to the latest version. This can be overridden with `RAILPACK_PACKAGES`.

## Configuration

Railpack will build your Gleam application as an Erlang shipment using `gleam export erlang-shipment`, and run it using `./build/erlang-shipment/entrypoint.sh run`. By default, the source tree is not available in the final container.
