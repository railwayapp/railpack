---
title: Mise Configuration
description: How to customize your image using Mise configuration
---

Railpack is built on top of [Mise](https://mise.jdx.dev/). You can use the various mise configuration options
to customize the Railpack-generated image. For instance, you can set environment variables, allow precompiled
ruby versions, and add additional utilities like `jq` to your image all through the mise configuration toml.

## Philosophy

* We use the latest mise version. There is automated tooling setup to ensure
  the mise version on the latest Railpack version is no more than a couple weeks
  out of date.
* Railpack assumes the default Mise configuration options. For instance, we won't
  opt-in users to precompiled ruby ahead of when mise has scheduled it to become default.
* Railpack generates global mise configuration based on analyizing the application
  source code. However, this global configuration is set in `/etc/mise/config.toml`
  so it can easily be overwritten in your application.

## Default Settings

Railpack sets the following mise settings by default in the generated
`/etc/mise/config.toml`. These can be overridden in your own `mise.toml`.

| Setting | Value | Reason |
|---------|-------|--------|
| `paranoid` | `true` | Enforces HTTPS and stricter security validation |
| `trusted_config_paths` | `["/app"]` | Trusts app config files to avoid warnings during build |
| `idiomatic_version_file_enable_tools` | *(language list)* | Auto-reads version files like `.node-version`, `.python-version`, etc. |
| `minimum_release_age` | `"14d"` | Mise will omit language, package manager, etc versions released in the last two weeks. You can override this settings in your application's mise configuration. |
| `node.verify` | `false` | Skips asset signature verification for Node, since recently released versions may not yet have a public key |

## Customization

Use a mise configuration file or environment variables to customize mise in the Railpack-generated container.
Configuration files are generally a better idea.

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
