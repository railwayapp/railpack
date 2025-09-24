#!/usr/bin/env zsh

cd "${0:A:h}"
pnpx create-react-router@latest . --yes --overwrite --no-git-init
