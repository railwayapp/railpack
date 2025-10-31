---
title: PHP
description: Building PHP applications with Railpack
---

Railpack can automatically build and deploy PHP applications with FrankenPHP, a
modern and efficient PHP application server.

## Detection

Your project will be detected as a PHP application if any of these conditions
are met:

- An `index.php` file exists in the root directory
- A `composer.json` file exists in the root directory

## Versions

The PHP version is determined in the following order:

- Read from the `composer.json` file
- Defaults to `8.4`

Only PHP 8.2 and above are supported.

## Configuration

Railpack will configure [FrankenPHP](https://frankenphp.dev/) for your
application. For Laravel applications, the document root is set to the `public`
directory.

### Config Variables

| Variable                   | Description                                         | Example            |
| -------------------------- | --------------------------------------------------- | ------------------ |
| `RAILPACK_PHP_ROOT_DIR`    | Override the document root                          | `/app/public`      |
| `RAILPACK_PHP_EXTENSIONS`  | Additional PHP extensions to install                | `gd,imagick,redis` |
| `RAILPACK_SKIP_MIGRATIONS` | Disable running Laravel migrations (default: false) | `true`             |

### Custom Configuration

Railpack uses default
[Caddyfile](https://github.com/railwayapp/railpack/blob/main/core/providers/php/Caddyfile)
and
[php.ini](https://github.com/railwayapp/railpack/blob/main/core/providers/php/php.ini)
configuration files. You can override these by placing your own versions in your
project root:

- `/Caddyfile` - Custom Caddy server configuration
- `/php.ini` - Custom PHP configuration

### Startup Process

The application is started using a
[start-container.sh](https://github.com/railwayapp/railpack/blob/main/core/providers/php/start-container.sh)
script that:

- For Laravel applications:
  - Runs database migrations and seeding (enabled by default, can be disabled with `RAILPACK_SKIP_MIGRATIONS`)
  - Creates storage symlinks
  - Optimizes the application
- Starts the FrankenPHP server using the Caddyfile configuration

You can customize the startup process by placing your own `start-container.sh`
in the project root.

### PHP Extensions

PHP extensions are automatically installed based on:

- Requirements specified in `composer.json` (e.g., `ext-redis`)
- Extensions listed in the `RAILPACK_PHP_EXTENSIONS` environment variable

Example `composer.json` with required extensions:

```json
{
  "require": {
    "php": ">=8.2",
    "ext-pgsql": "*",
    "ext-redis": "*"
  }
}
```

## Laravel Support

Laravel applications are detected by the presence of an `artisan` file. When
detected:

- The document root is set to the `/app/public` directory
- Storage directory permissions are set to be writable
- Composer dependencies are installed
- Artisan caches are optimized at build time:
  - Configuration cache
  - Event cache
  - Route cache
  - View cache

## Node.js Integration

If a `package.json` file is detected in your PHP project:

- Node.js will be installed
- NPM dependencies will be installed
- Build scripts defined in `package.json` will be executed
- Development dependencies will be pruned in the final image

This is particularly useful for Laravel applications that use frontend
frameworks like Vue.js or React.

You can see the [node docs](/languages/node) for information on how to configure
node.

## BuildKit Caching

The PHP provider will cache `/opt/cache/composer` under the cache key `composer`.