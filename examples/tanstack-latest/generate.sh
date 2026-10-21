#!/usr/bin/env zsh

cd "${0:A:h}"
npx @tanstack/cli@latest create . --yes --force --no-git
