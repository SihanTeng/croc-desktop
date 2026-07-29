import { describe, expect, it } from "vitest";
import { decodeDataUrlText, maxTextPreview, previewKind } from "./filepreview";

describe("previewKind", () => {
  it("maps media extensions to their kind", () => {
    expect(previewKind("photo.PNG")).toBe("image");
    expect(previewKind("clip.mp4")).toBe("video");
    expect(previewKind("song.flac")).toBe("audio");
    expect(previewKind("notes.md")).toBe("text");
  });

  it("uses the last extension and ignores case", () => {
    expect(previewKind("archive.tar.gz")).toBeNull();
    expect(previewKind("IMG_0001.JPEG")).toBe("image");
  });

  it("returns null for unknown or missing extensions", () => {
    expect(previewKind("setup.exe")).toBeNull();
    expect(previewKind("Makefile")).toBeNull();
  });

  it("handles names in folders", () => {
    expect(previewKind("docs/report.txt")).toBe("text");
  });
});

describe("decodeDataUrlText", () => {
  it("decodes base64 data URLs", () => {
    const url = "data:text/plain;base64," + btoa("hello croc");
    expect(decodeDataUrlText(url)).toBe("hello croc");
  });

  it("decodes multibyte UTF-8 correctly", () => {
    const bytes = new TextEncoder().encode("hello 🐊 ünïcodé");
    const b64 = btoa(String.fromCharCode(...bytes));
    expect(decodeDataUrlText(`data:text/plain;charset=utf-8;base64,${b64}`)).toBe(
      "hello 🐊 ünïcodé"
    );
  });

  it("returns empty string for malformed input", () => {
    expect(decodeDataUrlText("not-a-data-url")).toBe("");
  });
});

describe("maxTextPreview", () => {
  it("is a quarter megabyte", () => {
    expect(maxTextPreview).toBe(256 * 1024);
  });
});
