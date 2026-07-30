import { expect, test } from "@playwright/test";
import { freshAppPage } from "./helpers.mjs";

test("drop zone: native file/folder drops land in the send list", async ({ page }) => {
  await freshAppPage(page);

  // the drop zone opts in as a Wails drop target
  const target = await page.evaluate(
    () => document.querySelector(".dropzone") && true
  );
  expect(target).toBe(true);

  // the native GTK drop surfaces as this backend event; simulate it with a
  // file and a folder
  await page.evaluate(() => {
    window.runtime.EventsEmit("files:dropped", ["/tmp/report.pdf", "/home/admin/Documents"]);
  });

  const names = page.locator(".file-list:not(.history-files) .file-name");
  await expect(names).toHaveCount(2);
  const sendList = page.locator(".file-list:not(.history-files)");
  await expect(sendList).toContainText("/tmp/report.pdf");
  await expect(sendList).toContainText("/home/admin/Documents");

  // dropping the same paths again does not duplicate them
  await page.evaluate(() => {
    window.runtime.EventsEmit("files:dropped", ["/tmp/report.pdf", "/home/admin/Documents"]);
  });
  await expect(names).toHaveCount(2);
});
