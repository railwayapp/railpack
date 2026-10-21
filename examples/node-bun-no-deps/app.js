// Integration test assertion: Railpack should omit Node for Bun-only apps with no dependencies.
// We shell out to `which -a` instead of Bun.which() because `bun run` injects a temporary node
// shim at /tmp/bun-node-*/node for npm script compatibility that Bun.which() sees but is not on PATH.
const result = Bun.spawnSync({
  cmd: ["sh", "-lc", "which -a node 2>/dev/null || true"],
  stdout: "pipe",
});

const nodePaths = result.stdout
  .toString()
  .trim()
  .split("\n")
  .filter(Boolean)
  // Defensive: which -a does not return the bun shim today, but filter it if that ever changes.
  .filter((path) => !path.startsWith("/tmp/bun-node-"));

if (nodePaths.length > 0) {
  console.error(`node binary should not exist: ${nodePaths.join(", ")}`);
  process.exit(1);
}

const bunVersion = Bun.version.split(".")[0];
console.log(`hi from bun without a lock - using Bun major version ${bunVersion}`);
console.log("bun runtime without node binary");