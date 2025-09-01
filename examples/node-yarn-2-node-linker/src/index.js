const cowsay = require("cowsay");

function getYarnVersion() {
  const ua = process.env.npm_config_user_agent || "";
  return /\byarn\/([^\s]+)/.exec(ua)?.[1] ?? null; // e.g. "1.22.22", "3.6.1"
}

console.log(
  cowsay.say({
    text: `Hello from Yarn ${getYarnVersion()}`,
  })
);
