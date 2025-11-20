---
title: C/C++
description: Building C/C++ applications with Railpack
---

Railpack builds and deploys CMake- and Meson-based C/C++ applications with zero configuration.

## Detection

Your project will be detected as a C/C++ application if a `CMakeLists.txt` file or a `meson.build` file exists in the root directory.

## Versions

The latest versions of CMake (or Meson) and Ninja will be installed during build.

## Configuration

Railpack will build your application into a build directory at `/build`, and run the executable in that directory whose name matches the name of your project's root directory. The source tree will not be available in the final container by default, only the build directory.
