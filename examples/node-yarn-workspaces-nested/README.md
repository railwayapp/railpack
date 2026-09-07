# node-yarn-workspaces-nested

Reproduces [#463](https://github.com/railwayapp/railpack/issues/463): a Yarn
Berry monorepo with a deeply-nested workspace package causes a phantom duplicate
directory inside the container, making `yarn install` fail.

## Structure

The workspace glob is `packages/**/*` (recursive), with a package at
`packages/prisma/timescaledb/`. The intermediate directory `packages/prisma/`
happens to be named `prisma`, which matches the `prisma` entry in
`SupportingInstallFiles`.

## The bug

`SupportingInstallFiles` returns both:

- `packages/prisma/timescaledb/package.json` (matched as a file)
- `packages/prisma` (matched as a directory named `prisma`)

This produces two copy commands in the build plan:

1. `copy packages/prisma/timescaledb/package.json` → creates
   `/app/packages/prisma/timescaledb/`
2. `copy packages/prisma → packages/prisma` → because BuildKit uses
   `CopyDirContentsOnly: false`, the directory `prisma/` is placed *inside*
   `/app/packages/prisma/`, creating the phantom
   `/app/packages/prisma/prisma/timescaledb/`

Yarn then sees two workspaces with the same name and fails:

```
Internal Error: Duplicate workspace name @wildmetrics/prisma-timescale:
  /app/packages/prisma/timescaledb conflicts with
  /app/packages/prisma/prisma/timescaledb
```
