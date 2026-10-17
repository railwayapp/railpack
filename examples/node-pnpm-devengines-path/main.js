const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const sharp = require("sharp");

async function main() {
  const pnpm = execFileSync("pnpm", ["--version"], { encoding: "utf8" }).trim();
  assert.match(pnpm, /^(11|12)\./);
  console.log(`runtime pnpm: ${pnpm}`);
  const { info } = await sharp({
    create: { width: 1, height: 1, channels: 3, background: "white" }
  }).png().toBuffer({ resolveWithObject: true });
  assert.equal(info.width, 1);
  assert.equal(info.height, 1);
  console.log("sharp native operation succeeded");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
