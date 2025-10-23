const puppeteer = require("puppeteer");

(async () => {
  console.log("Starting Puppeteer");
  const browser = await puppeteer.launch({
    headless: true,
    args: ["--no-sandbox"],
  });
  const version = await browser.version();
  console.log(`Chrome version: ${version}`);
  console.log("Creating Page");
  const page = await browser.newPage();
  console.log("Navigating to hackernews");
  await page.goto("https://news.ycombinator.com", {
    waitUntil: "networkidle2",
  });

  await browser.close();
  console.log("Hello from puppeteer");
})();
