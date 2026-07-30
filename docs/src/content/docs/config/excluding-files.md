---
title: Excluding Files
description: How to exclude files from your build using .dockerignore
---

Railpack supports the standard `.dockerignore` file syntax to exclude files
from your build. For a complete reference on the syntax, please refer to the
[official Docker documentation](https://docs.docker.com/build/building/context/#dockerignore-files).

## How .dockerignore Works

When you use a `.dockerignore` file, Railpack parses it and passes all
patterns (including negations) to BuildKit in a single `exclude` list.
BuildKit's native pattern matching handles both exclusions and negations
(lines starting with `!`).

### Example

Given a `.dockerignore` file:

```
**/node_modules
.env
*.log
!important.log
```

When you run `railpack build --show-plan`, you'll see this gets converted to:

```json
{
  "exclude": [
    "**/node_modules",
    ".env",
    "*.log",
    "!important.log"
  ],
  "steps": [...],
  "deploy": {...}
}
```

## Using railpack.json

You can also specify `exclude` patterns directly in your `railpack.json`
configuration file instead of (or in addition to) using `.dockerignore`.
Negation patterns (starting with `!`) can be included in the `exclude` array:

```json
{
  "exclude": [
    "**/node_modules",
    "**/.venv",
    "*.log",
    ".env",
    "!important.log"
  ],
  "deploy": {
    "startCommand": "node server.js"
  }
}
```

This gives you more control and allows you to manage all build configuration in
one place. If both `.dockerignore` and `railpack.json` exclude patterns are
present, they are merged together.

## Default Behavior

If no `.dockerignore` file is present in your project root, Railpack defaults
to including **all files** in the build context.

## Common Patterns

### Excluding Nested Directories

A common gotcha: patterns without wildcards only match at the root level.

```
# Only excludes /node_modules (root level)
node_modules

# Excludes node_modules at ANY nesting level
**/node_modules
```

For example, if you have this structure:
```
project/
├── node_modules/      ← excluded by both patterns
└── web/
    └── node_modules/  ← only excluded by **/node_modules
```

Use `**/` prefix to match directories at any depth. This applies to all
directory names: `**/node_modules`, `**/.venv`, `**/vendor`, etc.

## Best Practices

### Local Development Artifacts

If you run builds locally, exclude local environment folders
like `node_modules`, `.venv`, or `vendor` at all nesting levels using the
`**/` prefix.

### Secrets and Metadata

You should exclude sensitive files and version control metadata to keep your
image clean, small, and secure.

It's recommend to add `.env`, any encrypted secrets, `.vscode`, `.github`, and
anything not required when running in production.

Here's a great [.dockerignore](https://configs.sh/dockerignore/) generator to
use as a starting point.
