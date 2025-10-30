---
title: ROS
description: Building ROS packages with Railpack
---

Railpack builds ROS packages as Docker images with zero configuration.

## Detection

Your project will be detected as a ROS package if a `package.xml` file exists in the root directory.

## Versions

The official ROS image will be used for the builds, with all components pulled from the official APT repository. This differs from other providers, which use Mise for most, if not all, dependencies.

## Configuration

Railpack will run `rosdep update` and `rosdep install` to install your application's dependencies, and build it with `colcon build`. If a launch file is specified with `RAILPACK_ROS_LAUNCH_FILE` (relative to `launch/`), that file will be launched as part of the start process; otherwise, the first launch file found (alphabetically) in `launch/` will be launched. If no launch file is specified or found, the start command will set up the ROS environment and do nothing further.

### Variables
| Variable | Description | Example | Default |
| -------- | ----------- | ------- | ------- |
| `RAILPACK_ROS_VERSION` | The version of ROS to use. | `rolling` | `kilted` |
| `RAILPACK_ROS_LAUNCH_FILE` | The launch file to use, relative to `launch/`. | `launch-everything.xml` | (none) |
