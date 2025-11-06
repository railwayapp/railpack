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

## Config Variables

| Variable                | Description                              | Example     |
| ----------------------- | ---------------------------------------- | ----------- |
| `RAILPACK_SHELL_SCRIPT` | Specify a custom shell script to execute | `deploy.sh` |
