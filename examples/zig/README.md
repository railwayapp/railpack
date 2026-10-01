# zig

A Zig executable built with `zig build`. The binary that `zig build` installs
into `zig-out/bin` is named after `.name` in `build.zig.zon`.

This example is generated from the template that ships with the compiler
(`zig init`) rather than hand-written. Run `./generate.sh` to regenerate it
after a Zig upgrade; the diff then shows exactly what changed upstream.

Only `src/main.zig` differs from the generated output: it prints the Zig
version to stdout. The integration test asserts on stdout, and
`std.debug.print` writes to stderr.
