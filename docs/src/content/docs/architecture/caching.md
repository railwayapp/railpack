---
title: Caching
description: Understanding Railpack's caching mechanisms
---

Railpack uses both BuildKit layer and mount caches to speed up successive
builds.

## Layer Cache

Railpack takes advantage of BuildKit's layer cache and avoids busting the cache
when possible. Cache busting events are defined in a granular way as part of the
[steps commands list](/architecture/overview/#build-step). These include:

- Copying files from the local context to the build context
- Changing environment variables
- Adding new generated files to the build context
- Executing shell commands in the build context

## Mount Cache

The [BuildKit mount
cache](https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md#run---mounttypecache)
is used to save the contents of a directory from the build context between
builds. This is useful for speeding up commands that download or compile assets
(e.g. npm install). The directory **does not** appear in the final image.

Caches are defined in the build plan and referenced by steps that need them.
Each cache has a type and a directory:

### Cache Types

- `shared`: Multiple builds can use this cache simultaneously (used for package
  manager caches)
- `locked`: Only one build can use this cache at a time (used for apt caches to
  prevent concurrent package installations)

Caches are shared across all steps that reference them. This is useful for
common caches such as the apt-cache or apt-lists.
