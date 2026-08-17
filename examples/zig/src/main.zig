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
