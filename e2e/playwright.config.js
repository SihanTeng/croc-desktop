import { defineConfig } from "@playwright/test";

// One backend, one transfer at a time: tests run strictly serially.
export default defineConfig({
  timeout: 120_000,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  globalTeardown: "./teardown.mjs",
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : "list",
  use: {
    // locally use the system Chrome (no bundled-browser download needed);
    // CI installs the bundled chromium and runs with the default channel
    channel: process.env.PLAYWRIGHT_CHANNEL || (process.env.CI ? undefined : "chrome"),
    viewport: { width: 1024, height: 720 },
    actionTimeout: 30_000,
  },
});
