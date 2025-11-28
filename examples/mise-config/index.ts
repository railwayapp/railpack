import { execSync } from "child_process";

function printVersion(cmd: string, label: string) {
  try {
    const version = execSync(cmd, { encoding: "utf8" }).trim();
    console.log(`${label} version: ${version}`);
  } catch (err) {
    throw new Error(`Failed to get ${label} version: ${err}`);
  }
}

printVersion("node --version", "Node.js");
printVersion("bun --version", "Bun");
printVersion("python --version", "Python");
printVersion("go version", "Go");
