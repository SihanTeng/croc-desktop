// Theme application: resolves the Settings theme ("system" | "light" |
// "dark") onto the document root so the CSS token overrides in
// tokens.css ([data-theme="dark"]) take effect.

import { Settings } from "./api";

const media = window.matchMedia("(prefers-color-scheme: dark)");
let current: Settings["theme"] = "system";

export function applyTheme(theme: Settings["theme"]) {
  current = theme;
  const dark = theme === "dark" || (theme === "system" && media.matches);
  document.documentElement.dataset.theme = dark ? "dark" : "light";
}

// re-resolve on OS theme changes while in "system" mode
media.addEventListener("change", () => applyTheme(current));
