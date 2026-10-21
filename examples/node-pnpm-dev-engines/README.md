`package.json` declares Node and pnpm only through `devEngines`. There is
no `packageManager` field and no lockfile.

Detection should still select pnpm 10.4.1.
