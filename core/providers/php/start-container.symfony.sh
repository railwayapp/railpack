#!/bin/bash

set -e

if [ "$RAILPACK_SKIP_MIGRATIONS" != "true" ]; then
  # Only when Doctrine Migrations is available (webapp pack / doctrine-migrations-bundle)
  if php bin/console list --raw 2>/dev/null | grep -q '^doctrine:migrations:migrate'; then
    echo "Running Doctrine migrations ..."
    php bin/console doctrine:migrations:migrate --no-interaction --allow-no-migration
  fi
fi

mkdir -p var/cache var/log
chmod -R a+rw var 2>/dev/null || true
php bin/console cache:warmup

echo "Starting Symfony server ..."

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
