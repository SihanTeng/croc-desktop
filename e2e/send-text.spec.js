import { expect, test } from "@playwright/test";
import { freshAppPage, spawnPeer } from "./helpers.mjs";

test("send text: code card, copy command, peer receives, send again", async ({ page }) => {
  await freshAppPage(page);

  await page.click(".segment:has-text('Text')");
  await page.fill(".text-input", "e2e hello from playwright");
  // Ctrl+Enter starts the send from the textarea
  await page.click(".text-input");
  await page.keyboard.press("Control+Enter");

  await expect(page.locator(".code-phrase")).toBeVisible();
  const code1 = (await page.textContent(".code-phrase")).trim();
  await expect(page.locator(".code-qr")).toBeVisible();
  await expect(page.locator("text=Waiting for recipient — share the code")).toBeVisible();

  // copy-command button gives inline feedback
  await page.click("button:has-text('Copy command')");
  await expect(page.locator("button:has-text('Copied ✓')").first()).toBeVisible();

  // a CLI-level peer receives the text
  const recv = spawnPeer("recv", [code1, ".e2e-tmp/recv-text"], { waitFor: "RECEIVE OK" });
  await expect(page.locator(".status-ok")).toBeVisible({ timeout: 90_000 });
  await recv.waitReady;

  // "send again" re-sends the same text under a fresh code
  await page.click("button:has-text('Send same text again')");
  await expect(page.locator(".code-phrase")).not.toHaveText(code1);
  const code2 = (await page.textContent(".code-phrase")).trim();
  const recv2 = spawnPeer("recv", [code2, ".e2e-tmp/recv-text"], { waitFor: "RECEIVE OK" });
  await expect(page.locator(".status-ok")).toBeVisible({ timeout: 90_000 });
  await recv2.waitReady;
});
