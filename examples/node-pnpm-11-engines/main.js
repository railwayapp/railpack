const { execSync } = require('child_process');

function getPnpmVersion() {
  const ua = process.env.npm_config_user_agent || "";
  return /\bpnpm\/([^\s]+)/.exec(ua)?.[1] ?? null;
}

const nodeVersion = process.version;
console.log(`Node.js version: ${nodeVersion}`);

const pnpmVersionFromUA = getPnpmVersion();
if (pnpmVersionFromUA) {
  console.log(`PNPM version (from user agent): v${pnpmVersionFromUA}`);
} else {
  console.log('PNPM version (from user agent): not detected');
}

try {
  const pnpmVersion = execSync('pnpm --version', { encoding: 'utf8' }).trim();
  console.log(`PNPM version (from CLI): v${pnpmVersion}`);
} catch (error) {
  console.log('PNPM version (from CLI): not available');
}
