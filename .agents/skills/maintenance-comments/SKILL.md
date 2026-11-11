---
name: maintenance-comments
description: >
  Find `// MAINTENANCE` comments and carry out each instruction after checking
  current upstream support, installable versions, and package names. Use when
  asked to handle maintenance comments, refresh provider defaults, or run a
  maintenance pass.
---

# Maintenance comments

Search the repository for the marker `// MAINTENANCE`. The rest of that line is the instruction. A URL on the following comment line is part of the source to check. Comments that merely say to keep something aligned, without the marker, are not a second queue. Update an unmarked copy when it cites the same upstream list as a marked one, so the two copies do not diverge.

Research each instruction against a primary source before editing. Record the date of the check. Do not change a value that still matches the instruction.

## Version pins

Read the wording on the line.

- "Oldest version continuing to receive security updates": if the pinned release is still inside its security window, leave it. Do not walk it backward to an older supported release. A newer pin that still receives patches is an intentional floor; lowering it changes the runtime for new apps, and `--previous` does not protect a first build. If the pin's security window has ended, move it to the oldest release that is still supported, not the newest. Keep the same version precision the constant already uses.
- "Latest version continuing to receive security updates": move the pin to the newest release the vendor still supports.

Confirm the new value with both the vendor policy and the installer this repo actually runs (mise, or the image tag source). A release that endoflife.date lists but mise cannot resolve must not become the default. Use the newest release the installer can install, and say so. Check `mise latest <tool>@<version>` before editing.

Paired defaults must stay compatible. For Elixir and Erlang, the default Erlang major has to sit inside the OTP range of the default Elixir minor. `getCompatibleErlangVersion` returns the newest OTP major that Elixir minor supports. Add a case only for a released minor. Compare the switch to the compatibility table on the current Elixir hexdocs, not an old versioned page.

An image suffix such as FrankenPHP's Debian tag is current when the latest image tags still publish it. Do not invent a newer Debian codename that the image has not moved to.

## Lockfiles and package lists

Add a lockfile-version mapping only when a new lockfile version exists. Leave an existing mapping on the oldest release that writes that format. A later major that reuses the format is not a new mapping.

For native library lists, use the upstream list for the runtime distro in `images/debian/runtime`. On Debian 13, Playwright's `debian13` Chromium list in `nativeDeps.ts` is the source of truth for that copy. `t64` names are the real packages. The unsuffixed names are often virtual `Provides` and still install, but the upstream Debian 13 list uses the real names. Check `Packages` for the runtime suite before renaming: a name that is neither a real package nor a virtual provide will fail `apt-get install`.

For a Python or Node package map, change an entry when that package's current install documentation changed. Add a package the map does not already mention only when the instruction covers that class of package and the requirement is documented. Skip a system package that would pull the distro interpreter when the build uses a mise-installed interpreter (`python3-dev` is the usual case).

A larger change needs a reason tied to the instruction: a new ecosystem package, retargeting a lockfile major, switching a concrete pin to a floating channel, or downgrading a pin that still receives security updates. Do not make that change on inference alone.

## After editing

Update docs that state a default you changed, and tests or examples that assert that default. Leave examples that pin their own version through `mise.toml`, `.tool-versions`, `packageManager`, `engines`, a lockfile mapping, or a toolchain file.

Plan snapshots omit the generated mise toml (`[generated-mise-toml]`), so a version-pin change often does not touch them. Apt package lists do.

Validate with `mise run check`, then `mise run test`. Refresh snapshots with `UPDATE_SNAPS=true mise run test`. `mise run test-update-snapshots` also `git add`s snapshot directories; do not run it when git writes are disallowed. Review the snapshot diff and restore unrelated drift, including Caddy versions, FrankenPHP patch tags, and mise image tags. Run the integration tests for examples whose resolved version or installed packages the edit changes. Do not run the full integration suite.
