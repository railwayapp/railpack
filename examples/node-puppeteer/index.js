const puppeteer = require("puppeteer");

// Cross-platform CI builds linux/amd64 under QEMU on arm runners; Chrome
// startup and navigation need more than Puppeteer's 30s defaults.
const qemuFriendlyTimeoutMs = 180_000;

(async () => {
  console.log("Starting Puppeteer");
  const browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox"],
    timeout: qemuFriendlyTimeoutMs,
  });
  const version = await browser.version();
  console.log(`Chrome version: ${version}`);
  console.log("Creating Page");
  const page = await browser.newPage();
  page.setDefaultTimeout(qemuFriendlyTimeoutMs);
  console.log("Navigating to example.com");
  await page.goto("https://example.com", {
    waitUntil: "networkidle2",
    timeout: qemuFriendlyTimeoutMs,
  });

  await browser.close();
  console.log("Hello from puppeteer");
})();
