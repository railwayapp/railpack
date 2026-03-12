---
title: CLI Reference
description: Complete reference for the Railpack CLI commands
---

Complete reference documentation for all Railpack CLI commands.

## Common Options

The following options are available across multiple commands:

| Flag                    | Description                                                                                                                |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `--env`                 | Environment variables to set. Format: `KEY=VALUE`                                                                          |
| `--previous`            | Versions of packages used for previous builds. These versions will be used instead of the defaults. Format: `NAME@VERSION` |
| `--build-cmd`           | Build command to use                                                                                                       |
| `--start-cmd`           | Start command to use                                                                                                       |
| `--config-file`         | Path to config file to use                                                                                                 |
| `--error-missing-start` | Error if no start command is found. Enabled by default on Railway.                                                         |

## Commands

### build

Builds a container image from a project directory using BuildKit.

**Usage:**

```bash
railpack build [options] DIRECTORY
```

**Options:**

| Flag          | Description                                           | Default |
| ------------- | ----------------------------------------------------- | ------- |
| `--name`      | Name of the image to build                            |         |
| `--output`    | Output the final filesystem to a local directory      |         |
| `--platform`  | Platform to build for (e.g. linux/amd64, linux/arm64) |         |
| `--progress`  | BuildKit progress output mode (auto, plain, tty)      | `auto`  |
| `--show-plan` | Show the build plan before building                   | `false` |
| `--cache-key` | Unique id to prefix to cache keys                     |         |

### prepare

Generates build configuration files without performing the actual build. This is
useful for platforms that want to:

- Build with a custom frontend and need to save the build plan to a
  `railpack-plan.json` file
- Log the Railpack pretty output to stdout
- Save the additional build information for later use

**Usage:**

```bash
railpack prepare [options] DIRECTORY
```

**Options:**

| Flag         | Description                                           |
| ------------ | ----------------------------------------------------- |
| `--plan-out` | Output file for the JSON serialized build plan        |
| `--info-out` | Output file for the JSON serialized build result info |

### plan

Analyzes a directory and outputs the build plan that would be used.

**Usage:**

```bash
railpack plan [options] DIRECTORY
```

**Options:**

| Flag          | Description                   |
| ------------- | ----------------------------- |
| `--out`, `-o` | Output file name for the plan |

### info

Provides detailed information about a project's detected configuration,
dependencies, and build requirements.

**Usage:**

```bash
railpack info [options] DIRECTORY
```

**Options:**

| Flag       | Description                  | Default  |
| ---------- | ---------------------------- | -------- |
| `--format` | Output format (pretty, json) | `pretty` |
| `--out`    | Output file name             |          |

### schema

Outputs the JSON schema for Railpack configuration files, used by IDEs for
autocompletion and validation.

**Usage:**

```bash
railpack schema
```

### frontend

Starts the BuildKit GRPC frontend server for internal build system use.

**Usage:**

```bash
railpack frontend
```

## Global Options

These options can be used with any command:

| Flag              | Description              |
| ----------------- | ------------------------ |
| `--help`, `-h`    | Show help information    |
| `--version`, `-v` | Show version information |
| `--verbose`       | Enable verbose logging   |
