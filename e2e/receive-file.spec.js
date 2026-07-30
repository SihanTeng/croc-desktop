import { expect, test } from "@playwright/test";
import path from "node:path";
import { freshAppPage, spawnPeer, uniqueCode, writeFixture, PIXEL_PNG } from "./helpers.mjs";

test("receive a file by pasting a CLI hint; inline image preview", async ({ page }) => {
  await freshAppPage(page);

  const png = writeFixture("pixel.png", PIXEL_PNG);
  const code = uniqueCode();
  const sender = spawnPeer("send", [code, png], { waitFor: "SENDER READY" });
  await sender.waitReady;

  await page.click(".rail-item:has-text('Receive')");
  // the paste-anything input extracts the code from a CLI hint
  await page.fill(".code-input", `CROC_SECRET="${code}" croc`);
  await page.fill(".input-row .input:visible", path.resolve(".e2e-tmp/out"));
  await page.click(".content .btn-primary.btn-lg:visible");

  // extraction worked iff the transfer actually starts (peer code matched)
  await page.click("button:has-text('Accept')", { timeout: 30_000 });

  const img = page.locator("img.recv-media");
  await expect(img).toBeVisible({ timeout: 90_000 });
  await page.waitForFunction(() => {
    const el = document.querySelector("img.recv-media");
    return el && el.complete && el.naturalWidth > 0;
  });
  await expect(page.locator(".recv-files .file-name")).toContainText("pixel.png");
  await expect(page.locator(".status-ok")).toHaveText("Transfer complete.");
});
