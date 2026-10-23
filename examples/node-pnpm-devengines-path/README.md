# pnpm version / PATH mismatch through a symlink

Reproducer for [#786](https://github.com/railwayapp/railpack/issues/786).
The same project builds through its real directory but fails at
`pnpm add -g node-gyp` when the input directory is a symlink.
There is no PATH workaround in this example.

## Reproduce

Run from the repository root, with the development tools and BuildKit configured
as described in the [local development guide](../../docs/src/content/docs/guides/developing-locally.md).
The commands build images and download Node.js, pnpm, and dependencies.

```sh
repo_root="$(pwd -P)"
example="$repo_root/examples/node-pnpm-devengines-path"
mkdir -p "$repo_root/tmp"
repro_dir="$(mktemp -d "$repo_root/tmp/issue-786.XXXXXX")"
ln -s "$example" "$repro_dir/project"

# Passing control: real directory.
mise run cli -- --verbose build "$example" \
  --name railpack-issue-786-control --progress plain --show-plan
docker run --rm railpack-issue-786-control

# Failing case: same files, reached through a symlink.
mise run cli -- --verbose build "$repro_dir/project" \
  --name railpack-issue-786-symlink --progress plain --show-plan
```

The control should print `runtime pnpm: ...` and
`sharp native operation succeeded`. The symlink case should fail before project
dependency installation with:

```text
ERR_PNPM_GLOBAL_BIN_DIR_NOT_IN_PATH
The configured global bin directory "/opt/pnpm/bin" is not in PATH
```

`test.json` covers the passing real-directory control and executes a native
`sharp` operation. It does **not** create a symlink or automatically exercise
the failing case; run the commands above for that comparison.

## Inspect planning without building

Using the variables from above:

```sh
mise run cli -- prepare "$example" \
  --plan-out "$repro_dir/control-plan.json" \
  --info-out "$repro_dir/control-info.json"
mise run cli -- prepare "$repro_dir/project" \
  --plan-out "$repro_dir/symlink-plan.json" \
  --info-out "$repro_dir/symlink-info.json"

# Additional comparison corresponding to the report's attempted pnpm-11 pin.
mise run cli -- prepare "$repro_dir/project" \
  --config-file railpack-pnpm11.json \
  --plan-out "$repro_dir/pnpm11-plan.json" \
  --info-out "$repro_dir/pnpm11-info.json"
```

Compare the `packages:mise` step's generated tool versions and the `install`
step's PATH commands. Plan inspection alone does not establish a runtime failure.

## Observed results

Tested on Linux amd64 with rootless Podman 5.8.7 running BuildKit, against
Railpack commit `4bda480` and its bundled Mise 2026.10.11:

| Input | Planned pnpm | Install PATH addition | Build result |
| --- | --- | --- | --- |
| Real directory | 12.14.0 | `/opt/pnpm/bin` | Passed; application ran with pnpm 12.8.1 and exercised `sharp` |
| Symlink directory | 10.34.6 | `/opt/pnpm` | Mise selected pnpm 12.9.0 inside the install step; global-bin PATH error |

The requested ranges are intentional: `engines.pnpm` admits pnpm 10, while
`devEngines.packageManager` requires pnpm 11 or 12. `prepare` ensures the manifest
is copied into the install step, as in the original application. The lockfile
was generated with pnpm 12.8.1, not edited manually. Versions may change as the
ranges resolve against newer releases and Mise's release-age policy.

Plan-only comparisons with Railpack 0.39.0 / Mise 2026.8.16 also showed:

- Real directory: pnpm 12.14.0 with `/opt/pnpm/bin`.
- Symlink: pnpm 10.34.6 with `/opt/pnpm`.
- Symlink plus `railpack-pnpm11.json`: pnpm 11.28.1, still with `/opt/pnpm`.

The historical comparisons were not built. The full runtime failure above was
reproduced on current code, not just inferred from a plan.

## Suspected mechanism and limits

Mise reports the manifest's real path, whereas `GetMisePackageVersions` filters
sources using a string-prefix comparison against the supplied app path. Through
a symlink those paths differ, so the idiomatic `devEngines` selection is dropped
and planning falls back to `engines.pnpm`. Once copied to `/app`, Mise can see the
manifest and select a newer pnpm whose global executables require the `/bin`
subdirectory absent from the plan's PATH.

The pnpm-11 configuration is a separate plan-only comparison: the custom package
override is applied after the provider has selected its PATH.

This establishes a trigger for the reported error, not proof that the original
Dokploy deployment used a symlink. No product fix or PATH workaround is included.

## Validation

- `mise run check` passed.
- The real-directory build and native-module runtime check passed.
- The symlink build failed at the reported global `node-gyp` installation.
- `mise run test` added the example's plan snapshot, but failed on 16 existing
  snapshots resolving Caddy 2.11.7 instead of their recorded 2.11.6. Those existing
  snapshots were left unchanged.
