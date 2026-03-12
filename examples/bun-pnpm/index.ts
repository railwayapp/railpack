import { execSync } from "node:child_process";
import chalk from "chalk";

process.stdout.write(chalk.green("Bun v:") + " " + Bun.version + "\n");

const pnpmVersion = execSync("pnpm --version", { encoding: "utf8" }).trim();
process.stdout.write(chalk.green("pnpm v:") + " " + pnpmVersion + "\n");
