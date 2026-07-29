import { describe, expect, it } from "vitest";
import { formatBytes } from "./api";

describe("formatBytes", () => {
  it("formats zero and negatives as 0 B", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(-5)).toBe("0 B");
  });

  it("formats bytes without a fraction", () => {
    expect(formatBytes(512)).toBe("512 B");
  });

  it("formats KB/MB/GB with one decimal", () => {
    expect(formatBytes(1536)).toBe("1.5 KB");
    expect(formatBytes(1048576)).toBe("1.0 MB");
    expect(formatBytes(2500000)).toBe("2.5 MB");
    expect(formatBytes(1073741824)).toBe("1.1 GB");
  });

  it("rounds to whole numbers at 100+", () => {
    expect(formatBytes(153600)).toBe("154 KB");
  });

  it("climbs units correctly", () => {
    expect(formatBytes(999)).toBe("999 B");
    expect(formatBytes(1000)).toBe("1.0 KB");
    expect(formatBytes(999999)).toBe("1000 KB");
    expect(formatBytes(1e12)).toBe("1.0 TB");
  });
});
