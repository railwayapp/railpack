#!/usr/bin/env zsh
# Description: autogen this repo from the latest version of the RR template

cd "${0:A:h}"

pnpx create-react-router@latest . --yes --overwrite --no-git-init

# disable SSR for SPA build
sed -i '' 's/ssr: true/ssr: false/' react-router.config.ts
