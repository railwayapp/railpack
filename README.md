# Railpack

[![CI](https://github.com/railwayapp/railpack/actions/workflows/ci.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/ci.yml)
[![Run Tests](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml)

Railpack is a tool for building images from source code with minimal
configuration. It is the successor to [Nixpacks](https://nixpacks.com) and
incorporates many of the learnings from running Nixpacks in production at
[Railway](https://railway.com) for several years.

## Getting Started

Install Railpack:

```bash
curl -sSL https://raw.githubusercontent.com/railwayapp/railpack/main/install.sh | bash
```

Create a sample JavaScript application using Vite:

```bash
npm create vite@latest my-app -- --template react
cd my-app
```

Build a container image with Railpack:

```bash
railpack build . --name my-app
```

That's it! Railpack automatically detects your project type and generates an
optimized container image.

## Documentation

Full documentation for both operators and users is available at
[railpack.com](https://railpack.com).

## Contributing

Railpack is open source and open to contributions. See the
[CONTRIBUTING.md](CONTRIBUTING.md) file for more information.
