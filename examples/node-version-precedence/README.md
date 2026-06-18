`package.json` specifies `"engines": { "node": ">=18" }` while
`RAILPACK_NODE_VERSION=22` is set at deploy time.

The ENV var should win.
