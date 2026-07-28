---
title: Installation
description: How to install Railpack
---

Railpack is available as a CLI tool. The latest release is available [on
GitHub](https://github.com/railwayapp/railpack/releases).

The BuildKit frontend is available as a [Docker image on
GHCR](https://github.com/railwayapp/railpack/pkgs/container/railpack-frontend).

## Mise

We love mise, and you can install Railpack using mise:

```sh
mise use github:railwayapp/railpack@latest
```

## Curl

Download Railpack from GH releases and install automatically

```sh
curl -sSL https://railpack.com/install.sh | sh
```

You can also customize the version, destination, and other config options:

```sh
curl -sSL https://railpack.com/install.sh | RAILPACK_VERSION=0.2.3 sh -s -- --bin-dir ~/.local/bin
```

## GitHub Releases

Go to the [latest release](https://github.com/railwayapp/railpack/releases) and
download the `railpack` binary for your platform.

## From Source

```sh
git clone https://github.com/railwayapp/railpack.git
cd railpack
go build -o railpack ./cmd/...

./railpack --help
```

## Agent Skill

The repository includes an [Agent Skill](https://agentskills.io) that teaches
compatible coding agents how to configure `RAILPACK_*` variables, construct
`railpack.json`, and run builds locally.

Install it for the current project with the
[Skills CLI](https://skills.sh/docs/cli):

```sh
npx skills add railwayapp/railpack --skill railpack
```

## Supported Platforms

Linux and MacOS are supported.

Windows builds are generated but not officially supported. That being said, PRs are welcome to fix any Windows-specific bugs.

## Help

Need help? Check out our [Help page](/help) for support options.
