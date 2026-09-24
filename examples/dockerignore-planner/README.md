# dockerignore-planner

A small app whose `.dockerignore` drops files the planner would otherwise
discover and copy.

Path discovery walks the source directory, but BuildKit only ever receives the
filtered context. Without filtering, `docs/marketing/package.json` is globbed up
and copied, and the build fails with:

```
failed to compute cache key: "/docs/marketing/package.json" not found
```

`mise.toml` pins Node 22. `mise.local.toml` pins Node 20 and is excluded, which
is the usual local-override setup. The plan must copy `mise.toml`, leave
`mise.local.toml` out, and resolve Node from `mise.toml`.
