const { chromium } = require("playwright");

(async () => {
  console.log("Starting Playwright");
  const browser = await chromium.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  const version = browser.version();
  console.log(`Chromium version: ${version}`);
  console.log("Creating Page");
  const page = await browser.newPage();
  console.log("Navigating to hackernews");
  await page.goto("https://news.ycombinator.com", {
    waitUntil: "networkidle",
  });

  await browser.close();
  console.log("Hello from playwright");
})();
