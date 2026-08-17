const std = @import("std");
const builtin = @import("builtin");

pub fn main() !void {
    std.debug.print("Hello from Zig {s}\n", .{builtin.zig_version_string});
}
