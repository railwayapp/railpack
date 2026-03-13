---
title: Mise Configuration
description: Understanding how Railpack uses Mise and how to configure it
---

Railpack is built on top of [Mise](https://mise.jdx.dev/). This enables you to use the various mise configuration options
to customize both your development and production environments.

## Philosophy

Railpack does not globally opt-in to non-default Mise configurations for
specific languages. We trust the tool authors to define the most reliable
defaults for their respective language ecosystems.

For example, although precompiled Ruby binaries can offer faster build
times, Railpack follows the Mise default of building Ruby from source. We
avoid forcing these unconventional behaviors at a global level to maintain
long-term stability and compatibility.

## Customization

You can customize how mise operates when setting up your application by
using a mise configuration file or environment variables. Configuration
files are generally a better idea.

### Configuration Files

Railpack automatically detects mise configuration files and passes them
into the build. This includes:

- **Config files**: `mise.toml`, `.mise.toml`, `mise/config.toml`,
  `.mise/config.toml`, `.config/mise.toml`, `.config/mise/config.toml`,
  `.tool-versions`
- **Environment-specific configs**: `mise.*.toml`, `.mise.*.toml`,
  `.config/mise/conf.d/*.toml`
- **Idiomatic version files**: `.ruby-version`, `.python-version`,
  `.python-versions`, `.node-version`, `.nvmrc`, `.go-version`,
  `.java-version`, `.sdkmanrc`, `.bun-version`, `.yvmrc`
- **Lock files**: `mise.lock` files co-located with any detected
  `*.toml` config

### Example: Precompiled Ruby

To opt-in to non-default features like precompiled Ruby, add a
`mise.toml` to your repository:

```toml
[tools]
ruby = "3"

[settings]
ruby.compile = false
```
