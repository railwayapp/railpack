---
title: Static Sites
description: Deploy static websites with Railpack
---

Railpack can automatically build and set up a server for static sites that
require no build steps. The [Caddy](https://caddyserver.com/) server is used as
the underlying web server.

## Detection

Your project will be automatically detected as a static site if any of these conditions are met:

- A `Staticfile` configuration file exists in the root directory
- An `index.html` file exists in the root directory
- A `public` directory exists
- The `RAILPACK_STATIC_FILE_ROOT` environment variable is set

## Root Directory Resolution

The provider determines the root directory in this order:

1. `RAILPACK_STATIC_FILE_ROOT` environment variable if set
2. `root` directory specified in `Staticfile` if present
3. `public` directory if it exists
4. Current directory (`.`) if `index.html` exists in root

## Configuration

### Staticfile

You can configure the root directory with a `Staticfile` that contains the contents:

```yaml
# root directory to serve
root: dist
```

### Config Variables

| Variable                    | Description                 | Example  |
| --------------------------- | --------------------------- | -------- |
| `RAILPACK_STATIC_FILE_ROOT` | Override the root directory | `public` |

### Custom Caddyfile

Railpack uses a custom
[Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/staticfile/Caddyfile.template)
that is used to serve the static files. You can overwrite this file with your
own Caddyfile at the root of your project.
