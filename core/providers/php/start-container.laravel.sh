#!/bin/bash

set -e

if [ "$RAILPACK_SKIP_MIGRATIONS" != "true" ]; then
  # Run migrations and seeding
  # https://laravel.com/docs/12.x/migrations#forcing-migrations-to-run-in-production
  echo "Running migrations and seeding database ..."
  php artisan migrate --force
fi

# https://laravel.com/docs/12.x/filesystem#the-public-disk
php artisan storage:link
# https://laravel.com/docs/12.x/deployment#optimization
php artisan optimize:clear
php artisan optimize

echo "Starting Laravel server ..."

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
