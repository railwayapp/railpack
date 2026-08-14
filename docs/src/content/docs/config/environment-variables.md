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
| `RAILPACK_PACKAGES`            | Install additional Mise packages. In the format `pkg[@version]`. The version is optional; if not provided, the latest version is used. Allows list.                             |
| `RAILPACK_BUILD_APT_PACKAGES`  | Install additional Apt packages during build. Allows list.                                                                                                                      |
| `RAILPACK_DEPLOY_APT_PACKAGES` | Install additional Apt packages in the final image. Allows list.                                                                                                                |

Variables which allow a list use space-separated values. For example:

```sh
RAILPACK_PACKAGES="pipx:httpie jq@latest"
```

To configure more parts of the build, it is recommended to use a [config file](/config/file).

## Global Options

These environment variables affect the behavior of Railpack:

They are read directly from Railpack's process environment; do not pass them
with `--env`.

| Name                      | Description                                                                 |
| :------------------------ | :-------------------------------------------------------------------------- |
| `FORCE_COLOR`             | Force colored output even when not in a TTY                                 |
| `RAILPACK_VERBOSE`        | Enable verbose logging (equivalent to the `--verbose` flag)                 |
| `RAILPACK_DISABLE_CACHES` | Disable specific BuildKit cache keys, or `*` to disable all caches. Allows list. |
