This example verifies that a Bun-only app with no dependencies and no lockfile builds
and runs correctly, and that the final image does not include the `node` binary.

Railpack omits Node for Bun projects without declared dependencies, since they do not
need node-gyp during install. The app shells out to `which -a node` (rather than
`Bun.which()`) to assert that no Node binary is on PATH in the built container.

Right now the cases where Node can be omitted are very limited, but we hope to expand
these over time and further expand the complexity of this example.

Related: https://github.com/railwayapp/railpack/issues/233