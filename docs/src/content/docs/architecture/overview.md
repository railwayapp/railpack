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

### Input Ordering and Layering

Each step (and the deploy step) assembles its filesystem from an ordered list
of inputs. When inputs reference the same path, ordering decides which one
wins.

- **Authoring order, not the graph.** Input lists are built by appending in a
  fixed sequence as providers run and as config is applied. The build graph is
  _derived from_ these inputs (each step-reference becomes an edge) — never the
  reverse. The order you see in the plan JSON is the source of truth.
- **Deterministic.** Any map-based sources (such as config steps and packages)
  are sorted by key before iteration, so Go's randomized map iteration never
  affects ordering. The same input always produces the same plan.
- **Base is pinned first.** The first input must be a full, unfiltered state
  (the foundation the others stack on). Inputs with `include`/`exclude` filters
  cannot be first.
- **Last wins.** Copy and merge operations are emitted strictly in input order,
  so for overlapping paths the later input takes precedence. To change which
  input wins, reorder the list — do not rely on graph topology.
- **Filters only affect the final image.** The `include`/`exclude` paths on an
  input control what is copied into the layer (e.g. the slim runtime image).
  They do not restrict what is available _during_ a step's execution, since a
  step's first input always brings the full filesystem of its source.

Separately, the build graph's topological sort decides _when_ each step's
state is computed (parents before children) so BuildKit can run independent
steps in parallel. This is independent of input ordering and does not change
the contents of the final image.

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

## Docker Images

Each Railpack binary has the builder and runtime image tags baked in at
compile time, derived from a single pinned mise version in
`core/mise/version.txt`:

- **Builder image** (`ghcr.io/railwayapp/railpack-builder:mise-<version>`):
  used during the build process. Contains mise, common languages, and build
  tools. Not included in the final image.
- **Runtime image** (`ghcr.io/railwayapp/railpack-runtime:mise-<version>`):
  a minimal Debian image used as the base for the final output image.

Both images include the `en_US.UTF-8` locale (generated at image build time).
Railpack does not set `LANG` or `LC_ALL` by default; set them yourself if your
application needs a UTF-8 locale (for example Python or Ruby apps that call
`locale.setlocale`).

Because the image tags are pinned to the mise version, upgrading Railpack
automatically uses the corresponding builder and runtime images. There is no
`latest` tag ambiguity — a given binary always references the same images.

If you want to use a specific builder or runtime image, you can customize the image
references in your `railpack.json`.

## Config

The build plan can be customized through [environment
variables](/config/environment-variables) (typically prefixed with `RAILPACK_`)
or through a [configuration file](/config/file). The configuration is applied to
the generate context after the providers have run.
