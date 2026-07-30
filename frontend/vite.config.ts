import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    watch: {
      // inotify misses bursts of rapid consecutive writes on this setup
      // (stale modules served after HMR); polling is cheap at this repo size
      usePolling: true,
      interval: 300,
    },
  },
});
