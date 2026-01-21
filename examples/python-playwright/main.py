from playwright.sync_api import sync_playwright

print("Starting Playwright")
with sync_playwright() as p:
    browser = p.chromium.launch(headless=True, args=["--no-sandbox"])
    version = browser.version
    print(f"Chromium version: {version}")
    print("Creating Page")
    page = browser.new_page()
    print("Navigating to hackernews")
    page.goto("https://news.ycombinator.com", wait_until="networkidle")
    browser.close()

print("Hello from playwright")
