import { execSync } from "node:child_process";

process.stdout.write("hello from Bun " + Bun.version + "\n");

const nodeVersion = execSync("node -v").toString().trim();
process.stdout.write("Node " + nodeVersion.replace(/^v/, "") + "\n");
