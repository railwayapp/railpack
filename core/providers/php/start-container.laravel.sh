#!/bin/bash

set -e

if [ "$RAILPACK_SKIP_MIGRATIONS" != "true" ]; then
  # Run migrations and seeding
  echo "Running migrations and seeding database ..."
  php artisan migrate --force
fi

php artisan storage:link
php artisan optimize:clear
php artisan optimize

echo "Starting Laravel server ..."

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
