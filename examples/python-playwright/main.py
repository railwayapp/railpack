from playwright.sync_api import sync_playwright

print("Starting Playwright")
with sync_playwright() as p:
    browser = p.chromium.launch(headless=True, args=["--no-sandbox"])
    version = browser.version
    print(f"Chromium version: {version}")
    print("Creating Page")
    page = browser.new_page()
    print("Navigating to example.com")
    page.goto("https://example.com", wait_until="networkidle")
    assert page.title() == "Example Domain"
    browser.close()

print("Hello from playwright")
