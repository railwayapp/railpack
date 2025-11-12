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

- Set via the `RAILPACK_RUBY_VERSION` environment variable
- Read from the `.ruby-version` file
- Read from the `Gemfile` file
- Read from mise-compatible version files (`.tool-versions`, `mise.toml`)
- Defaults to `3.4.6`

## Runtime Variables

These variables are available at runtime:

```sh
BUNDLE_GEMFILE="/app/Gemfile"
GEM_PATH="/usr/local/bundle"
GEM_HOME= "/usr/local/bundle"
MALLOC_ARENA_MAX="2"
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
- **Magick**: Installs `imagemagick`
- **Vips**: Installs `libvips-dev`
- **Charlock Holmes**: Installs `libicu-dev`
