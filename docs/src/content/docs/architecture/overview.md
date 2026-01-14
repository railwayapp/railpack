---
title: High Level Overview
description: Understanding Railpack's architecture and components
---

Railpack is split up into three main components:

- Core
  - The main logic that analyzes the app and generates the build plan
- BuildKit
  - Takes the build plan and generates [BuildKit
    LLB](https://github.com/moby/buildkit?tab=readme-ov-file#exploring-llb)
  - Starts a custom frontend or creates a BuildKit client to execute the build
    plan and generate an image
- CLI
  - The main entry point for Railpack

The core can be thought of as a _compiler_. The build plan that is generated is
independent from Docker, BuildKit, or any other tool that can be used to
generate an image. BuildKit is currently the primary _backend_, though the
architecture supports additional backends.

## Build Plan

The build plan is a JSON object that contains all the information necessary to
generate an image. Things that it includes are:

- Steps
  - List of build steps that execute commands and modify the filesystem
- Caches
  - Map of cache definitions that can be referenced by steps
- Secrets
  - List of secret names that are referenced by steps
- Deploy
  - Configuration for how the container runs, including:
    - Inputs: List of inputs for the deploy step
    - Start command: The command to run when the container starts
    - Variables: Environment variables available to the start command
    - Paths: Paths to prepend to the $PATH environment variable

### Build Step

A step is a group of commands that is executed sequentially in the build. Steps
explicitly define their inputs. These can be other steps, images, or local
files. The build graph is constructed in such a way that BuildKit will execute
non-dependent steps in parallel.

Steps contain:

- Name
  - Unique identifier for the step
- Inputs
  - List of inputs that define where the step gets its filesystem from:
    - Step input: Another step's output
    - Image input: A Docker image
    - Local input: Local files
- Commands
  - List of commands to run in the build:
    - Exec command: Run a shell command
    - Copy command: Copy files from source to destination
    - Path command: Add a directory to the global PATH
    - File command: Create a new file with optional permissions
- Secrets
  - List of secret names that this step uses
- Assets
  - Mapping of name to file contents referenced in file commands
- Variables
  - Mapping of name to variable values referenced in variable commands
- Caches
  - List of cache IDs available to all commands in this step

## Providers

Language support is managed through providers. Providers are typically
associated with a single language (e.g. node, python, php, etc.). A provider
will:

- Detect
  - Analyze the app and determine if it matches (e.g. the node provider will
    check for the presence of a `package.json` file)
- Build
  - Modifies the build context with all the steps, commands, caches, and
    everything that is needed to build for that language/framework

## Config

The build plan can be customized through [environment
variables](/config/environment-variables) (typically prefixed with `RAILPACK_`)
or through a [configuration file](/config/file). The configuration is applied to
the generate context after the providers have run.
