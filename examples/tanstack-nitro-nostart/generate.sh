#!/usr/bin/env zsh

cd "${0:A:h}"
npx @tanstack/cli@latest create . --deployment nitro --non-interactive --force --no-git --no-intent
