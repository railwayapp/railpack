#!/usr/bin/env zsh
set -euo pipefail

cd "${0:A:h}"
# Description: autogen this repo from the latest Symfony webapp skeleton

# composer create-project refuses a non-empty target directory (unlike tools
# such as create-react-router which support --overwrite). Wipe generated files
# and temporarily move this script + test.json out of the directory so
# create-project can run in place, then restore them.
find . -mindepth 1 -maxdepth 1 \
  ! -name generate.sh \
  ! -name test.json \
  -exec rm -rf {} +

# workspace tmp/ (gitignored); parent must exist for mktemp
mkdir -p "${0:A:h}/../../tmp"
meta_dir="$(mktemp -d "${0:A:h}/../../tmp/php-symfony-meta.XXXXXX")"
mv generate.sh test.json "$meta_dir/"

composer create-project symfony/skeleton:"8.*" . --no-interaction --prefer-dist
composer require webapp --no-interaction

# drop nested VCS if create-project initialized one
rm -rf .git

# Root railpack .gitignore ignores bare `bin` and `*.test` repo-wide.
# Re-include Symfony entrypoints so they can be committed with the example.
cat >> .gitignore <<'EOF'

# Railpack: un-ignore paths matched by the repo-root .gitignore
!/bin/
!/bin/**
!/.env.test
EOF

# Skeleton has no homepage route; add one for integration httpCheck.
mkdir -p src/Controller
cat > src/Controller/HomeController.php <<'EOF'
<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class HomeController
{
    #[Route('/', name: 'home')]
    public function index(): Response
    {
        return new Response('Welcome to Symfony');
    }
}
EOF

mv "$meta_dir"/generate.sh "$meta_dir"/test.json .
rmdir "$meta_dir"
