from playwright.sync_api import sync_playwright

def verify_blackjack(page):
    # Handle alerts
    page.on("dialog", lambda dialog: print(f"Alert: {dialog.message}"))

    page.goto("http://localhost:8080")

    # Check if Betting Controls are visible
    page.wait_for_selector("#betting-controls")

    # Check Initial Balance
    balance_text = page.locator("#player-balance").inner_text()
    if balance_text != "100":
        print(f"Error: Initial balance is {balance_text}, expected 100")

    # Place a Bet (5)
    page.fill("#bet-amount", "5")
    page.click("#start-btn")

    # Wait for game area
    try:
        page.wait_for_selector("#game-area", timeout=5000)
    except Exception as e:
        print(f"Timeout waiting for game area: {e}")

    # Check Balance (should be 95)
    page.screenshot(path="verification/betting_verification.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        try:
            verify_blackjack(page)
        finally:
            browser.close()
