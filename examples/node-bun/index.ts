import { execSync } from "node:child_process";

console.log("hello from Bun " + Bun.version);

const nodeVersion = execSync("node -v").toString().trim();
console.log("Node " + nodeVersion.replace(/^v/, ""));
