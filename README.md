# Railpack

[![CI](https://github.com/railwayapp/railpack/actions/workflows/ci.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/ci.yml)
[![Run Tests](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml)

Railpack is a tool for building images from source code with minimal
configuration. It is the successor to [Nixpacks](https://nixpacks.com) and
incorporates many of the learnings from running Nixpacks in production at
[Railway](https://railway.com) for several years.

## Getting Started

```bash
# Install mise (manages dev tools)
curl -sSL https://mise.run | sh

# Install Railpack
curl -sSL https://railpack.com/install.sh | sh

# Start BuildKit container
docker run --rm --privileged -d --name buildkit moby/buildkit

# Create a sample Vite + React app
npm create vite@latest my-app -- --template react
cd my-app

# Build container image
railpack build . --name my-app
```

Railpack automatically detects your project type and generates an optimized
container image.

**Note:** The above steps are for running Railpack locally to experiment and
test. If you deploy on a platform like [Railway](https://railway.com), Railpack
runs automatically when you push changes to your repository—no setup required.

## Documentation

Full documentation for both operators and users is available at
[railpack.com](https://railpack.com).

## Contributing

Railpack is open source and open to contributions. See the
[CONTRIBUTING.md](CONTRIBUTING.md) file for more information.
