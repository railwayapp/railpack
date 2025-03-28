---
title: Elixir
description: Building Elixir applications with Railpack
---

Railpack builds and deploys Elixir and Phoenix applications with zero configuration.

## Detection

Your project will be detected as a Elixir application if a `mix.exs` file exists in the root directory.

## Versions

The Elixir version is determined in the following order:

- Set via the `RAILPACK_ELIXIR_VERSION` environment variable
- Set via the `.elixir-version` file
- Detected from the `mix.exs` file
- Defaults to `1.18`

The OTP version is determined in the following order:

- Set via the `RAILPACK_ERLANG_VERSION` environment variable
- Set via the `.erlang-version` file
- Detected automatically from the resolved Elixir version

## Configuration

Railpack builds your Elixir application based on your project structure. The build process:

- Installs Elixir and Erlang
- Gets and compiles dependencies using `mix deps.get --only prod` and `mix deps.compile`
- If defined, deploys assets and ecto using `mix assets.deploy` and `mix ecto.deploy`
- Compiles a release for the project using `mix compile` and `mix release`
- Sets up the start command from your release binary

The selected file will be run with `/app/_build/prod/rel/{}/bin/{} start`.

### Config Variables

| Variable                  | Description                 | Example |
| ------------------------- | --------------------------- | ------- |
| `RAILPACK_ELIXIR_VERSION` | Override the Elixir version | `1.18`  |
| `RAILPACK_ERLANG_VERSION` | Override the Erlang version | `27.3`  |
