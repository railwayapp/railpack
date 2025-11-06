#!/usr/bin/env bun

/**
 * Development script to test building with the Railpack BuildKit frontend.
 *
 * This script validates the production BuildKit frontend flow by:
 * 1. Generating a build plan using `railpack plan`
 * 2. Invoking BuildKit with the frontend container image (ghcr.io/railwayapp/railpack:railpack-frontend)
 * 3. Loading the resulting image into Docker
 *
 * This is useful for testing the frontend integration locally before publishing,
 * as the frontend path is what production platforms use. The `railpack build` command
 * uses a different approach (direct BuildKit client) for simplicity.
 *
 * Example:
 *   bun scripts/run-with-frontend.ts examples/node-vite-react
 *   bun scripts/run-with-frontend.ts examples/node-vite-react --env NODE_ENV=production
 */

import { existsSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import crypto from "node:crypto";

const PLAN_FILE = "railpack-plan.json";
const FRONTEND_IMAGE = "ghcr.io/railwayapp/railpack:railpack-frontend";

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length === 0) {
  console.error("Please provide a directory path");
  process.exit(1);
}

const dir = args[0];
const envArgs = args.slice(1).filter((arg) => arg.startsWith("--env"));

// Create temp directory to save the build plan to
const randId = Math.random().toString(36).slice(2);
const tmpDir = join("/tmp", "railpack-" + randId);
const planDir = join(tmpDir, "plan");
mkdirSync(planDir, { recursive: true });

const cleanup = () => {
  if (existsSync(tmpDir)) {
    rmSync(tmpDir, { recursive: true, force: true });
  }
};

process.on("exit", cleanup);
process.on("SIGINT", () => {
  cleanup();
  process.exit();
});

// Generate build plan
console.log(`Generating build plan for ${dir}`);
const planResult = spawnSync(
  "go",
  ["run", "cmd/cli/main.go", "plan", dir, ...envArgs],
  {
    stdio: ["inherit", "pipe", "inherit"],
  }
);

if (planResult.status !== 0) {
  console.error("Failed to generate build plan");
  process.exit(1);
}

const planPath = join(planDir, PLAN_FILE);
writeFileSync(planPath, planResult.stdout);

// Parse all env vars so that we can use them as secrets
const envVars: Record<string, string> = {};
const secretArgs: string[] = [];

// Find all env args and their values
for (let i = 0; i < args.length; i++) {
  if (args[i] === "--env" && i + 1 < args.length) {
    const nameValue = args[i + 1];
    const [name, value] = nameValue.split("=");
    if (name && value) {
      envVars[name] = value;
      secretArgs.push(`--secret=id=${name},env=${name}`);
    }
    i++; // Skip the next argument since we've processed it
  }
}

// Pipe buildctl and docker load together
const buildctlArgs = [
  "build",
  `--local`,
  `context=${dir}`,
  `--local`,
  `dockerfile=${planDir}`,
  "--frontend=gateway.v0",
  "--opt",
  `source=${FRONTEND_IMAGE}`,
  "--output",
  "type=docker,name=test",
  ...secretArgs,
];

// Options that are passed to our custom frontend
const cacheKey = dir;
buildctlArgs.push("--opt", `cache-key=${cacheKey}`);

if (Object.keys(envVars).length > 0) {
  const secretsHash = crypto
    .createHash("sha256")
    .update(Object.values(envVars).sort().join(""))
    .digest("hex");
  buildctlArgs.push("--opt", `secrets-hash=${secretsHash}`);
}

console.log(`Executing buildctl\n  ${buildctlArgs.join(" ")}`);

const buildctl = spawnSync("buildctl", buildctlArgs, {
  stdio: ["inherit", "pipe", "inherit"],
  env: { ...process.env, ...envVars },
});

if (buildctl.status !== 0) {
  console.error("buildctl command failed");
  process.exit(1);
}

const dockerLoad = spawnSync("docker", ["load"], {
  input: buildctl.stdout,
  stdio: ["pipe", "inherit", "inherit"],
});

if (dockerLoad.status !== 0) {
  console.error("docker load failed");
  process.exit(1);
}
