# Node version precedence

Reproduces [railwayapp/railpack#591](https://github.com/railwayapp/railpack/issues/591).

## Problem

`package.json` specifies `"engines": { "node": ">=18" }` while
`RAILPACK_NODE_VERSION=22` is set at deploy time.

Per the [Node.js docs](https://railpack.com/languages/node), the environment
variable should take priority over `engines.node`. This example verifies that
precedence: Railpack must resolve Node 22 from `RAILPACK_NODE_VERSION`, not Node
18 from `package.json > engines > node`.

## Version priority

| Priority | Source | This example |
|----------|--------|--------------|
| 1 | `RAILPACK_NODE_VERSION` | `22` (set in `test.json`) |
| 2 | `engines.node` | `>=18` (in `package.json`) |

The integration test expects runtime output `Node version: v22`.