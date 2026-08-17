#!/usr/bin/env bash
#
# Regenerates this example from the canonical `zig init` template.
#
# The Zig compiler ships its own project template, so this example is derived
# from it rather than hand-written. Re-run after a Zig upgrade: the diff shows
# exactly what upstream changed.
#
# Only src/main.zig is not taken verbatim. The generated one pulls in the
# sample library module and demo tests; this example replaces it with a
# minimal program that prints the Zig version to *stdout* (the integration
# test asserts on stdout, and `std.debug.print` writes to stderr).
#
# Usage:  ./generate.sh
set -euo pipefail

cd "$(dirname "$0")"

command -v zig >/dev/null || { echo "zig not found in PATH" >&2; exit 1; }

PROJECT_NAME="zig_example"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# `zig init` derives the package, module and binary name from the directory
# it runs in, and refuses to touch a non-empty directory. Generate inside a
# fixed-name subdirectory so those names are deterministic.
mkdir -p "$tmp/$PROJECT_NAME"
( cd "$tmp/$PROJECT_NAME" && zig init >/dev/null 2>&1 )

# The fingerprint is a permanent package identity: Zig generates it once and
# it must never change afterwards. `zig init` mints a new one on every run,
# so carry the existing one over instead of committing a fresh identity.
if [ -f build.zig.zon ]; then
  existing_fingerprint="$(grep -oE '\.fingerprint = 0x[0-9a-f]+' build.zig.zon || true)"
else
  existing_fingerprint=""
fi

cp "$tmp/$PROJECT_NAME/build.zig" "$tmp/$PROJECT_NAME/build.zig.zon" .
cp "$tmp/$PROJECT_NAME/src/root.zig" src/root.zig

if [ -n "$existing_fingerprint" ]; then
  sed -i.bak -E "s/\.fingerprint = 0x[0-9a-f]+/${existing_fingerprint}/" build.zig.zon
  rm -f build.zig.zon.bak
fi

cat > src/main.zig <<'EOF'
const std = @import("std");
const builtin = @import("builtin");
const Io = std.Io;

pub fn main(init: std.process.Init) !void {
    // Stdout is for the actual output of the program. `std.debug.print`
    // writes to stderr, so it is deliberately not used here.
    var stdout_buffer: [64]u8 = undefined;
    var stdout_file_writer: Io.File.Writer = .init(.stdout(), init.io, &stdout_buffer);
    const stdout = &stdout_file_writer.interface;

    try stdout.print("Hello from Zig {s}\n", .{builtin.zig_version_string});
    try stdout.flush(); // Don't forget to flush!
}
EOF

echo "regenerated from 'zig init' ($(zig version))"
