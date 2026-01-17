---
title: Shell Scripts
description: Deploy applications using shell scripts with Railpack
---

Railpack can deploy applications that use shell scripts as their entry point.

## Detection

Your project will be automatically detected as a shell script application if any
of these conditions are met:

- A `start.sh` script exists in the root directory
- The `RAILPACK_SHELL_SCRIPT` environment variable is set to a valid script file

## Script File

Create a shell script in your project root (e.g., `start.sh`):

```bash
#!/bin/bash

echo "Hello world..."
```

## Shell Interpreter Detection

Railpack automatically detects which shell interpreter to use by reading the
shebang line in your script. The following shells are supported:

| Shell  | Shebang Example      | Notes                            |
| ------ | -------------------- | -------------------------------- |
| `bash` | `#!/bin/bash`        | Available in base image          |
| `sh`   | `#!/bin/sh`          | Available in base image          |
| `dash` | `#!/bin/dash`        | Available in base image (uses sh)|
| `zsh`  | `#!/bin/zsh`         | Automatically installed          |

If no shebang is present, `sh` is used as the default.

### Unsupported Shells

Non-POSIX shells like `fish`, `mksh`, and `ksh` cannot be automatically
detected and will fall back to `bash`. If you need to use these shells, you
may need to install them manually in your script or use a supported shell.

## Custom Installation

You can use the `RAILPACK_INSTALL_CMD` environment variable to run a custom installation command before the build step. This is useful for creating configuration files, downloading artifacts, or setting up the environment.

```bash
RAILPACK_INSTALL_CMD="mkdir -p config && echo 'production=true' > config/settings.conf"
```

You can also execute a custom script from your repository:

```bash
RAILPACK_INSTALL_CMD="bash setup.sh"
```

Files created during this step are available in the build step and the final image.

## Additional Packages

You can install additional packages using the `RAILPACK_PACKAGES` environment variable for Mise-supported tools, and `RAILPACK_BUILD_APT_PACKAGES` or `RAILPACK_DEPLOY_APT_PACKAGES` for system packages.

```bash
RAILPACK_PACKAGES="jq@latest python@3.11"
RAILPACK_DEPLOY_APT_PACKAGES="ffmpeg"
```

For more details, see the [environment variables documentation](/config/environment-variables).

## Config Variables

| Variable | Description | Example |
| :--- | :--- | :--- |
| `RAILPACK_SHELL_SCRIPT` | Specify a custom shell script to execute | `deploy.sh` |
