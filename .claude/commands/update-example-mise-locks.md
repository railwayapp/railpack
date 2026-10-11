Refresh mise lockfiles in `examples/` that already have one. If `$ARGUMENTS` contains `bump`, re-resolve fuzzy selectors with `mise lock --bump`. Otherwise refresh metadata for the versions already pinned.

Do not commit, push, or open a PR unless asked. Do not create a lockfile for an example that does not already have one. Do not edit `mise.toml` tool requests.

## 1. Find candidates

An example qualifies only when it has both a mise config and a mise lockfile. Package-manager locks (`uv.lock`, `Gemfile.lock`, `Cargo.lock`, `package-lock.json`) do not count.

```bash
find examples -type f \( \
  -name 'mise*.toml' -o -name '.mise*.toml' -o -name 'mise*.lock' -o -name '.mise*.lock' \
\) | sort
```

Group by example directory. Keep a directory when both groups are non-empty. `conf.d` fragments and `mise/config.toml` share the lockfile in the parent directory (`mise/mise.lock`, `mise.<env>.lock`, `mise.local.lock`), matching `lockfile_path_for_config` in https://github.com/jdx/mise/blob/main/src/lockfile.rs.

## 2. Refresh

Run from the example directory so that directory is the project config root. The repo root has its own `mise.toml`; `mise lock` there updates Railpack's lockfile, not the example.

Example configs are often untrusted. Pass `MISE_TRUSTED_CONFIG_PATHS` for the command. Do not run `mise trust`; that writes the user's trust database.

```bash
MISE_TRUSTED_CONFIG_PATHS="$PWD/examples/<name>" mise --cd examples/<name> lock
```

With `bump` in `$ARGUMENTS`:

```bash
MISE_TRUSTED_CONFIG_PATHS="$PWD/examples/<name>" mise --cd examples/<name> lock --bump
```

The repo `mise.toml` enables idiomatic version files, so a `.ruby-version` or `.python-version` in the example is locked too.

`mise lock` rewrites checksums and URLs for the pinned versions and leaves those versions in place. `--bump` re-resolves `latest`, `lts`, and prefixes such as `"3"` to the newest match and still does not edit the config. https://mise.jdx.dev/cli/lock.html

Use `--local` only when the tools live in a `.local.toml` and the lockfile is `mise.local.lock`. Do not pass `--global` or `--upgrade` unless asked.

## 3. Report

For each example: whether the lockfile changed, and any version moves (`old` → `new`). Say which examples had a mise config and no lockfile, and that they were left alone.
