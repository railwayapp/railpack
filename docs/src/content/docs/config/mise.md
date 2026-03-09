---
title: Mise Configuration
description: Understanding how Railpack uses Mise and how to configure it
---

Railpack is built on top of [Mise](https://mise.jdx.dev/) and adheres strictly to its established defaults. This ensures that your build environment is predictable and consistent with the broader Mise ecosystem.

## Philosophy

Railpack does not globally opt-in to non-default Mise configurations for specific languages. We trust the tool authors to define the most reliable defaults for their respective language ecosystems.

For example, although precompiled Ruby binaries can offer faster build times, Railpack follows the Mise default of building Ruby from source. We avoid forcing these unconventional behaviors at a global level to maintain long-term stability and compatibility.

## Customization

You can customize how mise operates when setting up your application by using a mise configuration file or environment variables. Configuration files are generally a better idea.

### Configuration Files

Railpack automatically detects all documented mise configuration files (`mise.toml`, `.tool-versions`, etc)/

### Example: Precompiled Ruby

To opt-in to non-default features like precompiled Ruby, add a `mise.toml` to your repository:

```toml
[settings]
# Use precompiled Ruby if supported by your mise version or plugin
ruby_precompiled = true
```
