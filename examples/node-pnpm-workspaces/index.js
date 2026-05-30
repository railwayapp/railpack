import { execSync } from "node:child_process";
import { greet as greetA } from "pkg-a";
import { greet as greetB } from "pkg-b";

function getPnpmVersion() {
  const ua = process.env.npm_config_user_agent || "";
  return /\bpnpm\/([^\s]+)/.exec(ua)?.[1] ?? null;
}

console.log("Node v:", process.version);

const pnpmVersionFromUA = getPnpmVersion();
if (pnpmVersionFromUA) {
  console.log(`PNPM version (from user agent): v${pnpmVersionFromUA}`);
} else {
  console.log("PNPM version (from user agent): not detected");
}

try {
  const pnpmVersion = execSync("pnpm --version", { encoding: "utf8" }).trim();
  console.log(`PNPM version (from CLI): v${pnpmVersion}`);
} catch {
  console.log("PNPM version (from CLI): not available");
}

console.log("Hello from pnpm workspaces");

console.log(greetA());
console.log(greetB());
