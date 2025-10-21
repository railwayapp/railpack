---
title: Rust
description: Building Rust applications with Railpack
---

Railpack builds and deploys Rust applications.

## Detection

Your project will be detected as a Rust application if any of these conditions are met:

- A `Cargo.toml` file is present

## Versions

The Rust version is determined in the following order:

- Any mise-supported version file (`mise.toml`, `.tool-versions`, etc).
- Read from the `toolchain.channel` field in the `rust-toolchain.toml` file
- Read from the `package.rust-version` field in the `Cargo.toml` file
- Read from the `.rust-version` or `rust-version.txt` file
- Set via the `RAILPACK_RUST_VERSION` environment variable
- Set via the `package.edition` field in the `Cargo.toml` file
- Defaults to `1.89`

## Runtime Variables

These variables are available at runtime:

```sh
ROCKET_ADDRESS="0.0.0.0"
```

## Configuration

Railpack builds your Rust application based on your project structure. The build process:

- Installs Rust and required system dependencies
- Installs package dependencies
- Compiles the application to a binary

The start command is:

```sh
./bin/<project-name>
```

### Config Variables

| Variable                   | Description                 | Example      |
| -------------------------- | --------------------------- | ------------ |
| `RAILPACK_RUST_VERSION`    | Override the Rust version   | `1.85.1`     |
