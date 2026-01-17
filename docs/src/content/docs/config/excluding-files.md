---
title: Excluding Files
description: How to exclude files from your build using .dockerignore
---

Railpack supports the standard `.dockerignore` file syntax to exclude files from your build. For a complete reference on the syntax, please refer to the [official Docker documentation](https://docs.docker.com/build/building/context/#dockerignore-files).

## Default Behavior

If no `.dockerignore` file is present in your project root, Railpack defaults to including **all files** in the build context.

### Local Development Artifacts

If you run builds locally, it is critical to exclude local environment folders like `node_modules`, `.venv`, or `vendor`. 

### Secrets and Metadata

You should exclude sensitive files and version control metadata to keep your image clean, small, and secure.

It's recommend to add `.env`, any encrypted secrets, `.vscode`, `.github`, and anything not required when running in production.

Here's a great [.dockerignore](https://configs.sh/dockerignore/) generator to use as a starting point.
