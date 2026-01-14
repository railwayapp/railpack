# Railpack

[![CI](https://github.com/railwayapp/railpack/actions/workflows/ci.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/ci.yml)
[![Run Tests](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml/badge.svg)](https://github.com/railwayapp/railpack/actions/workflows/run_tests.yml)

Railpack is a tool for building images from source code with minimal
configuration. It is the successor to [Nixpacks](https://nixpacks.com) and
incorporates many of the learnings from running Nixpacks in production at
[Railway](https://railway.com) for several years.

## Getting Started

```bash
# Install mise & railpack
curl -sSL https://mise.run | sh
mise install ubi:railwayapp/railpack@latest

# start BuildKit container & let railpack know about it
docker run --rm --privileged -d --name buildkit moby/buildkit
export BUILDKIT_HOST='docker-container://buildkit'

# create a Next.js app
npm create next-app@latest my-app
cd my-app

# build and run the app!
railpack build .
docker run -p 3000:3000 -it my-app
```

Railpack automatically detects the project type (Next.js, in this case, but many languages & frameworks are supported!) and generates an optimized
container image.

**Note:** The above steps are for running Railpack locally to experiment and
test. If you deploy on [Railway](https://railway.com), Railpack
runs automatically when you push changes to your repository.

## Documentation

Full documentation for both operators (platforms, like Railway) and users (developers using Railpack to build their applications) is available at
[railpack.com](https://railpack.com).

## Contributing

Railpack is open source and open to contributions. See the
[CONTRIBUTING.md](CONTRIBUTING.md) file for more information.
