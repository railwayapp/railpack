# RAILPACK_NODE_VERSION vs engines.node

Reproduces [railwayapp/railpack#591](https://github.com/railwayapp/railpack/issues/591).

## Problem

This example is the build directory itself — the same files Railway would see
when Root Directory is set to `/backend` in a monorepo. The subfolder layout
from the issue is not required to trigger the bug.

`package.json` specifies `"engines": { "node": ">=18" }` while
`RAILPACK_NODE_VERSION=22` is set at deploy time.

Per the [Node.js docs](https://railpack.com/languages/node), the environment
variable should take priority over `engines.node`. Instead, Railpack resolves
Node 18 from `package.json > engines > node`.

## Expected vs actual

| Source | Expected | Actual (bug) |
|--------|----------|--------------|
| `RAILPACK_NODE_VERSION=22` | Node 22 | Ignored |
| `engines.node: ">=18"` | Fallback | Node 18 selected |