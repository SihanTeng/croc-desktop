import { expect, test } from "@playwright/test";
import path from "node:path";
import { freshAppPage, spawnPeer, uniqueCode, writeFixture, PIXEL_PNG } from "./helpers.mjs";

test("declining an incoming transfer cancels it", async ({ page }) => {
  await freshAppPage(page);

  const png = writeFixture("nope.png", PIXEL_PNG);
  const code = uniqueCode();
  const sender = spawnPeer("send", [code, png], { waitFor: "SENDER READY" });
  await sender.waitReady;

  await page.click(".rail-item:has-text('Receive')");
  await page.fill(".code-input", code);
  await page.fill(".input-row .input:visible", path.resolve(".e2e-tmp/out"));
  await page.click(".content .btn-primary.btn-lg:visible");

  await page.click("button:has-text('Decline')", { timeout: 30_000 });
  // declining returns to the receive form and records a cancelled entry
  await expect(page.locator(".code-input:visible")).toBeVisible({ timeout: 15_000 });
  await page.click(".rail-item:has-text('History')");
  const first = page.locator(".history-item").first();
  await expect(first).toContainText("cancelled");
  await expect(first).toContainText("nope.png");
});

test("Esc cancels a waiting transfer", async ({ page }) => {
  await freshAppPage(page);

  await page.click(".rail-item:has-text('Receive')");
  await page.fill(".code-input", uniqueCode("nosender"));
  await page.fill(".input-row .input:visible", path.resolve(".e2e-tmp/out"));
  await page.click(".content .btn-primary.btn-lg:visible");
  // RelayView also renders a .card (hidden) — scope to the visible one
  await expect(page.locator(".card:visible")).toBeVisible();

  await page.keyboard.press("Escape");
  await expect(page.locator("text=Transfer cancelled.")).toBeVisible({ timeout: 15_000 });
});
