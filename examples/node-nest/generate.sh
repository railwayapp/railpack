#!/usr/bin/env zsh

root="${0:A:h}"
scratch="$root/tmp/node-nest"

rm -rf "$root/tmp"
mkdir -p "$scratch"

(
  cd "$scratch"
  npx -y @nestjs/cli@latest new app \
    --package-manager npm \
    --skip-git \
    --skip-install \
    --strict
)

# nest new creates a subdirectory; sync it into the example dir.
rsync -a --delete --exclude=generate.sh --exclude=test.json --exclude=tmp/ "$scratch/app/" "$root/"
rm -rf "$root/tmp"

cd "$root"
npm pkg set name=node-nest
# Run the compiled app directly: `nest start` would rebuild through the CLI on every boot.
npm pkg set scripts.start="node dist/main"
npm install --package-lock-only
