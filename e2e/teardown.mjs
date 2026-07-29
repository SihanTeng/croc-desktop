// Global teardown: leave the app's history/logs as we found them and drop
// the fixture workspace.
import { chromium } from "@playwright/test";
import { rmSync } from "node:fs";

export default async function globalTeardown() {
  rmSync(".e2e-tmp", { recursive: true, force: true });
  const url = process.env.E2E_APP_URL ?? "http://localhost:34115";
  const browser = await chromium.launch();
  try {
    const page = await browser.newPage();
    await page.goto(url);
    await page.waitForFunction(() => window.go?.main?.App, null, { timeout: 30_000 });
    await page.evaluate(() => {
      window.go.main.App.CancelTransfer();
      window.go.main.App.ClearHistory();
      window.go.main.App.ClearLogs();
    });
  } finally {
    await browser.close();
  }
}
