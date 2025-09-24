---
title: Installation
description: How to install Railpack
---

Railpack is available as a CLI tool. The latest release is available [on
GitHub](https://github.com/railwayapp/railpack/releases).

The BuildKit frontend is available as a [Docker image on
GHCR](https://github.com/railwayapp/railpack/pkgs/container/railpack-frontend).

## Mise

We love mise, and you can install railpack using mise:

```sh
mise use ubi:railwayapp/railpack@latest
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
