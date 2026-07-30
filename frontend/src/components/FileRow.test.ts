import { describe, expect, it } from "vitest";
import { fileIconKind } from "./FileRow";

describe("fileIconKind", () => {
  it("maps folders", () => {
    expect(fileIconKind("anything", true)).toBe("folder");
  });

  it("maps media types", () => {
    expect(fileIconKind("photo.JPG")).toBe("image");
    expect(fileIconKind("clip.mp4")).toBe("video");
    expect(fileIconKind("song.flac")).toBe("audio");
  });

  it("maps archives", () => {
    for (const n of ["a.zip", "b.tar.gz", "c.7z", "d.rar"]) {
      expect(fileIconKind(n)).toBe("archive");
    }
  });

  it("maps documents and text", () => {
    expect(fileIconKind("paper.pdf")).toBe("doc");
    expect(fileIconKind("notes.md")).toBe("text");
    expect(fileIconKind("data.csv")).toBe("text");
  });

  it("falls back to a generic file", () => {
    expect(fileIconKind("firmware.bin")).toBe("file");
    expect(fileIconKind("Makefile")).toBe("file");
  });
});
