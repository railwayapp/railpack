#!/usr/bin/env zsh
# Description: autogen this repo from the latest version of the RR template

cd "${0:A:h}"

pnpx create-react-router@latest . --yes --overwrite --no-git-init

# disable SSR for SPA build
sed -i '' 's/ssr: true/ssr: false/' react-router.config.ts

# Pin the build-time preview server to IPv4 until React Router connects to its bound address.
# https://github.com/remix-run/react-router/issues/15255
sed -i '' '/plugins: \[tailwindcss(), reactRouter()\],/a\
  // Keep React Router'\''s preview server and prerender request on the same address family.\
  preview: {\
    host: "127.0.0.1",\
  },
' vite.config.ts
