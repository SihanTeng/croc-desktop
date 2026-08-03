// Shared helpers for the E2E specs: app reset, peer spawning, fixtures.
import { spawn } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

export const APP_URL = process.env.E2E_APP_URL ?? "http://localhost:34115";

// a valid 1x1 transparent PNG
export const PIXEL_PNG = Buffer.from([
  0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
  0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
  0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
  0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
  0x42, 0x60, 0x82,
]);

export const WORK_DIR = path.resolve(".e2e-tmp");

export function fixtureDir() {
  mkdirSync(WORK_DIR, { recursive: true });
  return WORK_DIR;
}

export function writeFixture(name, data) {
  const p = path.join(fixtureDir(), name);
  writeFileSync(p, data);
  return p;
}

let counter = 0;
// unique codes avoid "room full" collisions on shared relays across runs
export function uniqueCode(prefix = "e2e") {
  counter += 1;
  return `${prefix}-${Date.now() % 100000}-${counter}`;
}

// open the app and clear any transfer state left by a previous test
export async function freshAppPage(page) {
  // Backend→frontend events (transfer:state, progress, …) travel over a
  // WebSocket that custom.js opens after page load; a send started before
  // the socket is open misses the one-shot "waiting" event and wedges the
  // UI. Wait for the server's "connected" log line before touching the UI.
  // (Absent in `wails3 dev`, where bound calls don't work anyway — fall
  // through after the timeout there.)
  const wsReady = page
    .waitForEvent("console", {
      predicate: (m) => m.text().includes("[Wails] Event WebSocket connected"),
      timeout: 10_000,
    })
    .catch(() => {});
  await page.goto(APP_URL);
  await wsReady;
  await page.waitForFunction(() => window.go?.main?.App, null, { timeout: 30_000 });
  // the backend cancels regardless of which browser page started the transfer
  await page.evaluate(() => window.go.main.App.CancelTransfer());
  // wait until the backend actually unwound (cancel is asynchronous)
  await page.waitForFunction(() => window.go.main.App.IsTransferRunning().then((r) => !r), null, {
    timeout: 15_000,
  });
  return page;
}

// spawn a test peer (see e2e/peers); routes through the sandbox relay when
// E2E_RELAY is set, otherwise the public croc relay.
// Flags go before positionals — Go's flag package stops at the first
// positional argument.
export function spawnPeer(kind, args, { waitFor } = {}) {
  const full = [];
  if (process.env.E2E_RELAY) {
    full.push("--relay", process.env.E2E_RELAY, "--pass", "pass123");
  }
  full.push(...args);
  const proc = spawn("go", ["run", "./peers", kind, ...full], { cwd: path.resolve(".") });
  let out = "";
  proc.stdout.on("data", (d) => {
    out += d;
    if (waitFor && out.includes(waitFor)) proc.emit("ready");
  });
  proc.stderr.on("data", (d) => {
    out += d;
  });
  proc.waitReady = waitFor
    ? new Promise((resolve, reject) => {
        const t = setTimeout(() => reject(new Error(`peer ${kind} never ready: ${out}`)), 60_000);
        proc.on("ready", () => {
          clearTimeout(t);
          resolve();
        });
        proc.on("exit", (code) => reject(new Error(`peer ${kind} exited ${code}: ${out}`)));
      })
    : Promise.resolve();
  return proc;
}
