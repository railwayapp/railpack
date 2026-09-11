#!/bin/bash

set -e

if [ "$IS_LARAVEL" = "true" ]; then
  # The config cache is compiled at build time, without the runtime environment,
  # so it can name a different database than the one this container is
  # configured for. Drop it before anything reads config. `optimize` below
  # rebuilds this cache, along with the event, route and view caches.
  php artisan config:clear

  if [ "$RAILPACK_SKIP_MIGRATIONS" != "true" ]; then
    # Run migrations and seeding
    echo "Running migrations and seeding database ..."
    php artisan migrate --force
  fi

  php artisan storage:link
  php artisan optimize

  echo "Starting Laravel server ..."
fi

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
