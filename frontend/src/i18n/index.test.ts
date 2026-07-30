import { beforeEach, describe, expect, it } from "vitest";
import { availableLanguages, getLanguage, resolveLanguage, setLanguage, t } from "./index";

describe("i18n", () => {
  beforeEach(() => setLanguage("en"));

  it("translates known keys", () => {
    expect(t("status.complete")).toBe("Transfer complete.");
  });

  it("interpolates variables", () => {
    expect(t("history.andMore", { n: 3 })).toBe("…and 3 more");
    expect(t("receive.savedTo", { dir: "/tmp" })).toBe("Saved to /tmp");
  });

  it("returns the key itself for unknown keys", () => {
    expect(t("no.such.key")).toBe("no.such.key");
  });

  it("translates into another locale and falls back to English for gaps", () => {
    setLanguage("zh-CN");
    expect(t("status.complete")).toBe("传输完成。");
    expect(t("history.andMore", { n: 2 })).toBe("…还有 2 项");
    expect(t("no.such.key")).toBe("no.such.key");
  });

  it("resolveLanguage maps explicit and unknown settings", () => {
    expect(resolveLanguage("fr")).toBe("fr");
    expect(resolveLanguage("xx-nothere")).toBe("en");
    expect(resolveLanguage("")).toBe("en");
  });

  it("resolveLanguage resolves system via navigator.language", () => {
    // test environment runs an English locale
    expect(availableLanguages).toContain(resolveLanguage("system"));
  });

  it("setLanguage applies the resolved locale", () => {
    setLanguage("ja");
    expect(getLanguage()).toBe("ja");
    setLanguage("xx-nothere");
    expect(getLanguage()).toBe("en");
  });

  it("ships the mainstream locales", () => {
    for (const l of ["en", "zh-CN", "zh-TW", "es", "fr", "de", "ja"]) {
      expect(availableLanguages).toContain(l);
    }
  });
});
