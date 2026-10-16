---
title: Ruby
description: Building Ruby applications with Railpack
---

Railpack builds and deploys Ruby applications with support for several
language-specific tools and frameworks.

## Detection

Your project is detected as a Ruby application if a `Gemfile` is present in
the root directory.

## Versions

The Ruby version is determined in the following order:

1. The `RAILPACK_RUBY_VERSION` environment variable
2. A [mise config file](/config/mise#configuration-files), such as `mise.toml`
   or `.tool-versions`
3. An [idiomatic version file](https://mise.jdx.dev/lang/ruby.html#version-files).
   Mise reads `.ruby-version` and a `ruby` directive in the `Gemfile`. The
   `Gemfile` directive wins when both set a version.
4. The `RUBY VERSION` section of `Gemfile.lock`, such as `ruby 3.1.2p20`.
   The patchlevel is ignored. Support for extracting this from the Gemfile.lock
   is deprecated.
5. Default to Ruby `3.4` as a last resort

## Runtime Variables

These variables are available at runtime:

```sh
BUNDLE_GEMFILE="/app/Gemfile"
GEM_PATH="/usr/local/bundle"
GEM_HOME="/usr/local/bundle"
MALLOC_ARENA_MAX="2"
LD_PRELOAD="libjemalloc.so.2"
```

## Configuration

Railpack builds your Ruby application based on your project structure. The build process:

- Installs Ruby and required system dependencies
- Installs project dependencies
- Configures the Ruby environment for production

The start command is determined by:

1. Framework-specific start command (see below)
2. `config/environment.rb` file
3. `config.ru` file
4. `Rakefile` file

### Config Variables

| Variable                   | Description                 | Example      |
| -------------------------- | --------------------------- | ------------ |
| `RAILPACK_RUBY_VERSION`    | Override the Ruby version   | `3.4.2`      |


## Framework Support

Railpack detects and configures caches and commands for popular frameworks:

### Rails

Railpack detects Rails projects by:

- Presence of `config/application.rb`

### Databases

Railpack automatically installs system dependencies for common databases:

- **PostgreSQL**: Installs `libpq-dev`
- **MySQL**: Installs `default-libmysqlclient-dev`
- **Magick**: Installs `libmagickwand-dev`
- **Vips**: Installs `libvips-dev`
- **Charlock Holmes**: Installs `libicu-dev`, `libxml2-dev`, `libxslt-dev`

### Node.js Integration

If a `package.json` file is detected in your Ruby project, or if the
`execjs` gem is used:

- Node.js will be installed automatically
- NPM dependencies will be installed if `package.json` exists
- Build scripts defined in `package.json` will be executed
- Development dependencies will be pruned in the final image

This is particularly useful for Rails applications with frontend assets.

### Bootsnap

Railpack automatically detects and optimizes applications using
[Bootsnap](https://github.com/Shopify/bootsnap):

- Runs `bundle exec bootsnap precompile --gemfile` during dependency
  installation
- Runs `bundle exec bootsnap precompile app/ lib/` during the build phase for
  Rails applications

### Asset Pipeline

Railpack detects Rails applications using asset pipeline gems:

- **Sprockets**: Runs `bundle exec rake assets:precompile` during build
- **Propshaft**: Runs `bundle exec rake assets:precompile` during build

Rails API-only applications without asset pipeline gems skip asset
compilation.

### Bundler Version

Railpack automatically detects the Bundler version from the `BUNDLED WITH`
section in your `Gemfile.lock` and installs that specific version.

### Precompiled Ruby

By default, Railpack follows the Mise default of building Ruby from
source. You can opt into precompiled Ruby binaries for faster build times
by adding a `mise.toml` to your repository:

```toml title="mise.toml"
[tools]
ruby = "3"

[settings]
ruby.compile = false
```

Note that precompiled binaries may not be available for all Ruby versions
or platforms.

### Performance Optimizations

Railpack includes several performance optimizations:

- **jemalloc**: Installs `libjemalloc2` and preloads it via `LD_PRELOAD`
  for improved memory allocation performance
- **YJIT**: For Ruby 3.2+, installs `rustc` and `cargo` required for YJIT
  compilation support

### Local Path Dependencies

If your Gemfile includes gems with local `path:` specifications, Railpack
will automatically copy those local directories during the build process.
