import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  plugins: [react(), wails("./bindings")],
  server: {
    host: "127.0.0.1",
    // Wails v3 dev mode defaults to 9245; keep 34115 if E2E/docs pin it via env
    port: Number(process.env.WAILS_VITE_PORT) || 34115,
    strictPort: true,
    watch: {
      // inotify misses bursts of rapid consecutive writes on this setup
      // (stale modules served after HMR); polling is cheap at this repo size
      usePolling: true,
      interval: 300,
    },
  },
});
