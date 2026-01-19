---
title: Excluding Files
description: How to exclude files from your build using .dockerignore
---

Railpack supports the standard `.dockerignore` file syntax to exclude files
from your build. For a complete reference on the syntax, please refer to the
[official Docker documentation](https://docs.docker.com/build/building/context/#dockerignore-files).

## How .dockerignore Works

When you use a `.dockerignore` file, Railpack parses it and converts it into
`exclude` and `include` patterns in the generated build plan. These patterns
are then applied when transferring your source code to BuildKit.

### Example

Given a `.dockerignore` file:

```
node_modules
.env
*.log
!important.log
```

When you run `railpack build --show-plan`, you'll see this gets converted to:

```json
{
  "exclude": [
    "node_modules",
    ".env",
    "*.log"
  ],
  "include": [
    "important.log"
  ],
  "steps": [...],
  "deploy": {...}
}
```

Note: Lines starting with `!` are negations that become include patterns.

## Using railpack.json

You can also specify `exclude` and `include` patterns directly in your
`railpack.json` configuration file instead of (or in addition to) using
`.dockerignore`:

```json
{
  "exclude": [
    "node_modules",
    ".venv",
    "*.log",
    ".env"
  ],
  "include": [
    "important.log"
  ],
  "deploy": {
    "startCommand": "node server.js"
  }
}
```

This gives you more control and allows you to manage all build configuration in
one place.

## Default Behavior

If no `.dockerignore` file is present in your project root, Railpack defaults
to including **all files** in the build context.

## Best Practices

### Local Development Artifacts

If you run builds locally, exclude local environment folders
like `node_modules`, `.venv`, or `vendor`.

### Secrets and Metadata

You should exclude sensitive files and version control metadata to keep your
image clean, small, and secure.

It's recommend to add `.env`, any encrypted secrets, `.vscode`, `.github`, and
anything not required when running in production.

Here's a great [.dockerignore](https://configs.sh/dockerignore/) generator to
use as a starting point.
