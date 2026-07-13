#!/bin/bash

set -e

# Start the FrankenPHP server
docker-php-entrypoint --config /Caddyfile --adapter caddyfile 2>&1
