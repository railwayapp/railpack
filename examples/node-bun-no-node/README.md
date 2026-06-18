This example verifies that a Bun-only app with no dependencies does not include
the `node` binary in the final image.

Right now the cases where this can occur are very limited, but we would hope to expand these
over time and further expand the complexity of this example.
