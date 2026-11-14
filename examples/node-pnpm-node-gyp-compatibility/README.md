# node-gyp compatibility with older Node versions

Verifies that pnpm uses a compatible node-gyp when building native dependencies
on an older Node version. Reproduces
[railpack#727](https://github.com/railwayapp/railpack/issues/727)
with the reported Node 20.20.2, pnpm 9.0.0, and better-sqlite3 12.11.1 versions.
The `.npmrc` forces compilation from source so a prebuilt binary cannot bypass
`node-gyp`. There is no project-local node-gyp dependency or override: the example
exercises the global version selected by Railpack.

From this directory, run:

```sh
mise run test-integration-cwd
```

Or, from the repository root:

```sh
mise run cli -- --verbose build --show-plan examples/node-pnpm-node-gyp-compatibility
```

Before compatibility selection was added, the integration build reproduced the
incompatibility on Linux arm64 with the Mise 2026.10.11 builder. The global
install selected node-gyp 13.1.0 and undici 8.12.2. The better-sqlite3 install script invoked that node-gyp
under Node 20.20.2 and failed during configuration:

```text
gyp info using node-gyp@13.1.0
gyp info using node@20.20.2 | linux | arm64
gyp ERR! stack TypeError: Cannot read properties of undefined (reading 'bind')
```

The stack points to `undici/lib/interceptor/decompress.js`. This is the same
Node 20/global node-gyp incompatibility as the issue, with a newer dependency
version and a different error. The install step now selects `node-gyp@12` for
Node 20.20.2. The same Linux arm64 test passes with node-gyp 12.4.0, including
compilation from source and a working SQLite query.
