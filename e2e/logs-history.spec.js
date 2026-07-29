import { expect, test } from "@playwright/test";
import { freshAppPage, spawnPeer } from "./helpers.mjs";

test("history and logs record a completed transfer; filters work", async ({ page }) => {
  await freshAppPage(page);

  // quick text send to a receiving peer
  await page.click(".segment:has-text('Text')");
  await page.fill(".text-input", "e2e history log entry");
  await page.click(".btn-primary.btn-lg:visible");
  await page.waitForSelector(".code-phrase");
  const code = (await page.textContent(".code-phrase")).trim();
  const recv = spawnPeer("recv", [code, ".e2e-tmp/recv-text"], { waitFor: "RECEIVE OK" });
  await expect(page.locator(".status-ok")).toBeVisible({ timeout: 90_000 });
  await recv.waitReady;

  await page.click(".rail-item:has-text('History')");
  const first = page.locator(".history-item").first();
  await expect(first).toBeVisible();
  await expect(first).toContainText("Sent");
  await expect(first).toContainText("completed");
  await expect(first).toContainText("e2e history log entry");

  await page.click(".rail-item:has-text('Logs')");
  await expect(page.locator(".log-row").first()).toBeVisible();
  await expect(page.locator(".logs-list")).toContainText("transfer complete");

  // level filter: no errors expected from a clean transfer
  await page.click(".segment:has-text('Error')");
  await expect(page.locator(".log-row")).toHaveCount(0);
  await page.click(".segment:has-text('All')");
  await expect(page.locator(".log-row").first()).toBeVisible();
});
