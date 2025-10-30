---
title: Meson
description: Building Meson applications with Railpack
---

Railpack builds and deploys Meson-based applications with zero configuration.

## Detection

Your project will be detected as a CMake application if a `meson.build` file exists in the root directory.

## Versions

The latest versions of CMake and Ninja will be installed during build.

## Configuration

Railpack will build your application into a build directory at `/build` with `meson setup /build` and `meson compile -C /build`, and run the executable in that directory whose name matches the name of your project's root directory. The source tree will not be available in the final container by default, only the build directory.
