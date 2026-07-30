---
title: Configuration Options
description: Ways to configure Railpack and how config affects build plans
---

Users can configure Railpack in a few different ways:

- [CLI flags](/reference/cli)
- [Environment variables](/config/environment-variables)
- [`railpack.json` config file](/config/file)
- [Mise configuration](/config/mise)

CLI flags, environment variables, and `railpack.json` are merged and then
applied to the generate context. Mise configuration files, such as
`mise.toml`, are also detected and used to configure tool versions, install
additional tools, set environment variables, and customize Mise behavior in
the generated image.

Everything that affects a part of the build plan _should_ be configurable.
Config affects the generate context rather than the plan itself as it allows
Railpack to perform optimizations after the config is applied. It also allows
the user config format to be abstracted at a higher level compared to the
relatively low level build plan schema.

## Configuration and Build Plans

`railpack.json` is a user-authored configuration file. It describes changes to
the generated build, but it is not a complete build plan.

Railpack combines `railpack.json`, CLI options, environment variables, Mise
configuration, and information detected from the application source to
generate a complete build plan. The `railpack prepare` command can serialize
this plan to a file commonly named `railpack-plan.json`.

The generated `railpack-plan.json` is a lower-level artifact consumed by the
[BuildKit frontend](/reference/frontend). It is the compiled output of the
planning process, not another `railpack.json` input. The two files have
different schemas and are not interchangeable.
