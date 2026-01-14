---
title: Dotnet
description: Building Dotnet applications with Railpack
---

Railpack builds and deploys Dotnet applications with zero configuration.

## Detection

Your project will be detected as a Dotnet application if a `*.csproj` file exists in the root directory.

## Versions

The Dotnet version is determined in the following order:

- Read `TargetFramework` from the first `*.csproj` file found in the project
- Read `version` from the `global.json` file in the project root
- Set via the `RAILPACK_DOTNET_VERSION` environment variable
- Defaults to `6.0.428`

## Configuration

Railpack builds your Dotnet application based on your project structure. The build process:

- Installs Dotnet SDK and Runtime
- Caches dependencies using `dotnet restore`
- Builds the project using `dotnet publish --no-restore -c Release -o out`
- Sets up the start command based on the publish output `./out/{project_name}`

### Config Variables

| Variable                  | Description                 | Example   |
| ------------------------- | --------------------------- | --------- |
| `RAILPACK_DOTNET_VERSION` | Override the Dotnet version | `6.0.428` |

### Runtime Packages

The `libicu-dev` package is installed to support internationalization in your Dotnet applications.

## Port Binding

Railpack automatically configures your application to listen on the port specified by the `PORT` environment variable (defaulting to 3000).

This is achieved by setting the `ASPNETCORE_URLS` environment variable in the start command:
`ASPNETCORE_URLS=http://0.0.0.0:${PORT:-3000}`.

This ensures the application listens on all network interfaces (`0.0.0.0`) rather than just `localhost`, making it accessible externally.
