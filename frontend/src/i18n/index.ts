// Tiny i18n layer. Locales are plain JSON files in ./locales — Vite picks up
// every file dropped there, so contributors add a language by adding one
// file, no code changes. Missing keys fall back to English.

/// <reference types="vite/client" />

import { useSyncExternalStore } from "react";
import en from "./locales/en.json";

type Messages = Record<string, string>;

const modules = import.meta.glob("./locales/*.json", { eager: true });

const locales: Record<string, Messages> = {};
for (const [path, mod] of Object.entries(modules)) {
  const lang = path.match(/\/([^/]+)\.json$/)?.[1];
  if (lang) locales[lang] = (mod as { default: Messages }).default;
}

/** all languages available in the bundle, English first */
export const availableLanguages = [
  "en",
  ...Object.keys(locales)
    .filter((l) => l !== "en")
    .sort(),
];

let current = "en";
const listeners = new Set<() => void>();

function emit() {
  for (const l of listeners) l();
}

/** resolve a Settings language ("system" | locale) to an available locale */
export function resolveLanguage(setting: string): string {
  if (setting !== "system" && setting !== "") {
    return locales[setting] ? setting : "en";
  }
  const nav = navigator.language || "en";
  if (locales[nav]) return nav;
  const base = nav.split("-")[0];
  if (base === "zh") {
    // Traditional-script regions fall to zh-TW, everything else zh-CN
    return /TW|HK|MO|Hant/i.test(nav) && locales["zh-TW"]
      ? "zh-TW"
      : locales["zh-CN"]
        ? "zh-CN"
        : "en";
  }
  const match = availableLanguages.find((l) => l === base || l.startsWith(base + "-"));
  return match ?? "en";
}

export function setLanguage(setting: string) {
  current = resolveLanguage(setting);
  emit();
}

export function getLanguage(): string {
  return current;
}

/** translate a key, interpolating {name} placeholders from vars */
export function t(key: string, vars?: Record<string, string | number>): string {
  let msg = locales[current]?.[key] ?? (en as Messages)[key] ?? key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      msg = msg.split(`{${k}}`).join(String(v));
    }
  }
  return msg;
}

/** React hook: t() that re-renders when the language changes */
export function useT() {
  useSyncExternalStore(
    (cb) => {
      listeners.add(cb);
      return () => listeners.delete(cb);
    },
    () => current
  );
  return t;
}
