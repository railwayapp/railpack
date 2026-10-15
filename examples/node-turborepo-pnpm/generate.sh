#!/usr/bin/env zsh

root="${0:A:h}"
scratch="$root/tmp/node-turborepo-pnpm"

rm -rf "$root/tmp"
mkdir -p "$root/tmp"

# create-turbo needs a new target directory; it installs, which writes pnpm-lock.yaml.
pnpm dlx create-turbo@latest "$scratch" --package-manager pnpm --no-git

rsync -a --delete \
  --exclude=generate.sh --exclude=test.json --exclude=tmp/ \
  --exclude=node_modules/ --exclude=.turbo/ \
  "$scratch/" "$root/"
rm -rf "$root/tmp"
