#!/usr/bin/env zsh

root="${0:A:h}"
scratch="$root/tmp/node-nx-next"

rm -rf "$scratch"
mkdir -p "$scratch"

(
  cd "$scratch"
  pnpm dlx create-nx-workspace@latest . \
    --preset=next \
    --workspaces \
    --appName=web \
    --nextAppDir=true \
    --pm=pnpm \
    --nxCloud=skip \
    --interactive=false \
    --skipGit \
    --e2eTestRunner=none \
    --unitTestRunner=none
)

# create-nx-workspace requires an empty target; sync into the example dir.
rsync -a --delete --exclude=generate.sh --exclude=tmp/ "$scratch/" "$root/"
rm -rf "$scratch"