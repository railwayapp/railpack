#!/bin/bash

set -e

# Ensure writable storage and cache directories exist before running artisan (e.g. when /app/storage is a mounted volume)
# https://laravel.com/docs/12.x/deployment#directory-permissions
mkdir -p storage/framework/{sessions,views,cache,testing} storage/logs bootstrap/cache
chmod -R a+rw storage bootstrap/cache

# Cached config bypasses env() and config/*.php, so clear any stale bootstrap/cache/config.php before running migrations
# https://laravel.com/docs/12.x/configuration#configuration-caching
# https://github.com/serversideup/docker-php/blob/main/src/common/etc/entrypoint.d/50-laravel-automations.sh
php artisan config:clear

if [ "$RAILPACK_SKIP_MIGRATIONS" != "true" ]; then
  # https://laravel.com/docs/12.x/migrations#forcing-migrations-to-run-in-production
  echo "Running migrations and seeding database ..."
  php artisan migrate --force
fi

# https://laravel.com/docs/12.x/filesystem#the-public-disk
php artisan storage:link
# Caches config, events, routes, and views in place; avoid optimize:clear as it also runs cache:clear and flushes app cache stores
# https://laravel.com/docs/12.x/deployment#optimization
# https://github.com/laravel/framework/blob/12.x/src/Illuminate/Foundation/Console/OptimizeClearCommand.php
php artisan optimize

echo "Starting Laravel server ..."

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
