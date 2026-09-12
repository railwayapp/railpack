# node-npm-dockerignore

A plain npm app that keeps a second, unrelated manifest under `docs/`, which
`.dockerignore` excludes.

Path discovery walks the source directory, but BuildKit only ever receives the
filtered context. Without filtering, `docs/marketing/package.json` is globbed up
and copied, and the build fails with:

```
failed to compute cache key: "/docs/marketing/package.json" not found
```
