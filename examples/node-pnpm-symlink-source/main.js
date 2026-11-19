const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");

// These pins must survive both planning through a symlink and execution through Mise's shims.
assert.equal(process.version, "v20.19.0");
const pnpmVersion = execFileSync("pnpm", ["--version"], { encoding: "utf8" }).trim();
assert.equal(pnpmVersion, "11.5.1");
console.log(`Node.js version: ${process.version}`);
console.log(`PNPM version: ${pnpmVersion}`);
