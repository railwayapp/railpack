Sync Railpack's project mise config filenames with the loader in current mise. Focus on `miseConfigFiles`, `miseConfigGlobs`, the detection list in the docs, and `TestGetSupportingMiseConfigFiles`. If `$ARGUMENTS` names a mise git ref, use that; otherwise use `main`.

Do not commit, push, or open a PR unless asked. Do not edit `miseIdiomaticFiles`.

## 1. Read what Railpack matches today

In `core/generate/mise_step_builder.go`:

- `miseConfigFiles` — exact paths, checked with `HasFile`
- `miseConfigGlobs` — doublestar patterns (`github.com/bmatcuk/doublestar/v4`), checked with `FindFiles`

Both slices are also passed to `ExcludedFiles` so a dockerignored config cannot change `mise list`. Keep one pair of slices. Do not add a third list.

`FindFiles` uses `doublestar.Glob`. `*` matches any run of non-`/` characters, including dots and empty. It does not cross `/`.

| Pattern | Matches | Does not match |
| --- | --- | --- |
| `mise.*.toml` | `mise.local.toml`, `mise.production.toml`, `mise.production.local.toml` | `mise.toml` |
| `.config/mise.*.toml` | `.config/mise.local.toml` | `.config/mise/mise.local.toml` |
| `mise/config.*.toml` | `mise/config.local.toml`, `mise/config.production.local.toml` | `mise/config.toml` |
| `mise/conf.d/*.toml` | `mise/conf.d/tools.toml`, `mise/conf.d/tools.production.toml` | `mise/conf.d/nested/skip.toml` |

Confirm an ambiguous pattern with `doublestar.Match` from a scratch program under `tmp/`, using the doublestar version in `go.mod`. Run it with `mise exec -- go run`.

## 2. Fetch what mise loads

Public list, in precedence order: https://mise.jdx.dev/configuration.html#mise-toml

That page leaves out config environments and some legacy paths. Also read:

- https://mise.jdx.dev/configuration/environments.html — `mise.<env>.toml`, `mise.<env>.local.toml`, the same suffixes on the other locations, and `conf.d` fragments
- `LOCAL_CONFIG_FILENAMES` and `env_config_patterns` in `src/config/mod.rs` on the chosen mise ref: https://github.com/jdx/mise/blob/main/src/config/mod.rs

The Rust lists are the loader. A path mise still loads belongs in Railpack even when the HTML page omits it.

Ignore these. They are not project config files at the app root:

- `~/.config/mise/**` and `/etc/mise/**`
- parent-directory search and `**/mise.toml` (mise loads a nested file when that directory is the cwd)
- `.miserc.toml` and `.config/miserc.toml` (they select `MISE_ENV`; they are not tool config)
- `MISE_OVERRIDE_CONFIG_FILENAMES`, `MISE_DEFAULT_CONFIG_FILENAME`, and `MISE_DEFAULT_TOOL_VERSIONS_FILENAME` (runtime overrides of the default names)

## 3. Classify

Build the set of project-root paths from `LOCAL_CONFIG_FILENAMES` plus every `env_config_patterns` entry, with the environment name replaced by `*`.

Collapse patterns that one glob already covers. `mise.local.toml`, `mise.<env>.toml`, and `mise.<env>.local.toml` are all `mise.*.toml`. The same collapse applies under `mise/config.`, `.mise/config.`, `.config/mise.`, and `.config/mise/config.`. Each `conf.d/*.<env>.toml` and `conf.d/*.<env>.local.toml` pattern is already `conf.d/*.toml`.

Put a name in `miseConfigFiles` only when no glob matches it. Put everything else in `miseConfigGlobs`. An exact name that a glob already matches does not get a second entry.

`.tool-versions` stays an exact file.

## 4. Edit

Update the two slices. Keep the doc URLs on `miseConfigGlobs`. A code comment may say why an exact name is absent from the public docs page. Do not describe that in user-facing docs.

Update the detection list in `docs/src/content/docs/config/mise.md` so it names the same files and globs. Keep it a flat list: config files, local and environment-specific configs, config fragments. No parenthetical about legacy names.

Extend `TestGetSupportingMiseConfigFiles` in `core/generate/mise_step_builder_test.go`. New positive paths go in `included`. Keep negatives for files mise does not load: `other.toml`, `mise/tasks/build.toml`, and a nested `conf.d` file.

## 5. Verify

```bash
mise run check
mise run test -- -run TestGetSupportingMiseConfigFiles
```

Do not run `mise run test-update-snapshots` unless a plan snapshot changes because an example contains a newly matched file. `tanstack-devdeps-spa` and `tanstack-nitro-nostart` already disagree with `core/mise/version.txt` on the image tag. Leave those snapshots alone.

## 6. Report

Paths added and removed, split into exact files and globs. Which mise ref you compared. Anything the HTML page omits that the Rust loader still has.
