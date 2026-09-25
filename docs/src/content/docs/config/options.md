---
title: Configuration Options
description: Ways to configure Railpack and how config affects build plans
---

Users can configure Railpack in a few different ways:

- [CLI flags](/reference/cli)
- [Environment variables](/config/environment-variables)
- [`railpack.json` config file](/config/file)
- [Mise configuration](/config/mise)

CLI flags, environment variables, and `railpack.json` are merged in precedence
order and then applied to the generate context. Mise configuration files are
used to configure tool versions, install additional tools, set environment
variables, and customize Mise behavior in the generated image.

Everything that affects a part of the build plan _should_ be configurable.
Config affects the generate context rather than the plan itself as it allows
Railpack to perform optimizations after the config is applied. It also allows
the user config format to be abstracted at a higher level compared to the
relatively low level build plan schema.

## Configuration Precedence

There are many configuration options which can be defined in multiple places.
Here is the order of precedence (highest wins):

1. CLI flags (`--start-cmd`, etc)
2. Environment variables passed with `--env` (`RAILPACK_START_CMD`, etc)
3. [`railpack.json`](/config/file)
4. [Procfile](/config/procfile) (only for the start command)
5. Provider and detected defaults

A couple things to note:

- An unset or empty value does not override a lower source.
- Build configuration from the environment must be passed with `--env`.
  Exported shell variables are not imported automatically.
- Array values from a higher source replace lower ones unless the value
includes `"..."` to extend the generated list. See [Array
Extending](/config/file#array-extending).
- Root secret names are an exception: names from `railpack.json` and `--env`
  are combined and deduplicated rather than replaced.
- Language versions (for example `RAILPACK_NODE_VERSION` versus `.nvmrc`) use a
per-language order documented on each language page.

### Mise Configuration Precedence

The build process generates a system-level `/etc/mise/config.toml`. You can
override these settings with a project-level `mise.toml` file.

## Configuration and Build Plans

`railpack.json` is a user-authored configuration file. It describes changes to
the generated build, but it is not a complete build plan.

Railpack combines `railpack.json`, CLI options, environment variables, Mise
configuration, and information detected from the application source to
generate a complete build plan. The `railpack prepare` command can serialize
this plan to a file commonly named `railpack-plan.json`.

The generated `railpack-plan.json` is a lower-level artifact consumed by the
[BuildKit frontend](/platforms/buildkit-frontend). It is the compiled output of
the planning process, not another `railpack.json` input. The two files have
different schemas and are not interchangeable.
