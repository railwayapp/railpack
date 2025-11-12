---
title: Environment Variables
description: Understanding environment variables in Railpack
---

Some parts of the build can be configured with environment variables. These are
often prefixed with `RAILPACK_`.

## Build Configuration

| Name                           | Description                                                                                                                                                                     |
| :----------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `RAILPACK_BUILD_CMD`           | Set the command to run for the build step. This overwrites any commands that come from providers                                                                                |
| `RAILPACK_INSTALL_CMD`         | Set the command to run for the install step. This overwrites any commands that come from providers. All files are copied to the root of the project before running the command. |
| `RAILPACK_START_CMD`           | Set the command to run when the container starts                                                                                                                                |
| `RAILPACK_PACKAGES`            | Install additional Mise packages. In the format `pkg@version`. The latest version is used if not provided.                                                                      |
| `RAILPACK_BUILD_APT_PACKAGES`  | Install additional Apt packages during build                                                                                                                                    |
| `RAILPACK_DEPLOY_APT_PACKAGES` | Install additional Apt packages in the final image                                                                                                                              |
| `RAILPACK_DISABLE_CACHES`      | Specify specific (space-separated) BuildKit cache keys to disable, or `*` to disable all caches. |

To configure more parts of the build, it is recommended to use a [config file](/config/file).

## Global Options

These environment variables affect the behavior of Railpack:

| Name          | Description                                 |
| :------------ | :------------------------------------------ |
| `FORCE_COLOR` | Force colored output even when not in a TTY |
