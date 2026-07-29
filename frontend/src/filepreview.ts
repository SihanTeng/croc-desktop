// Pure helpers for received-file previews, extracted for unit testing.

export type PreviewKind = "image" | "video" | "audio" | "text";

const previewExts: Record<string, PreviewKind> = {
  png: "image",
  jpg: "image",
  jpeg: "image",
  gif: "image",
  webp: "image",
  svg: "image",
  bmp: "image",
  avif: "image",
  ico: "image",
  mp4: "video",
  webm: "video",
  mkv: "video",
  mov: "video",
  m4v: "video",
  avi: "video",
  mp3: "audio",
  wav: "audio",
  ogg: "audio",
  flac: "audio",
  m4a: "audio",
  opus: "audio",
  txt: "text",
  md: "text",
  json: "text",
  log: "text",
  csv: "text",
};

// text previews larger than this would bog down the DOM; just list them
export const maxTextPreview = 256 * 1024;

export function previewKind(name: string): PreviewKind | null {
  const ext = name.split(".").pop()?.toLowerCase() ?? "";
  return previewExts[ext] ?? null;
}

export function decodeDataUrlText(dataUrl: string): string {
  const i = dataUrl.indexOf(",");
  if (i < 0) return "";
  const bytes = Uint8Array.from(atob(dataUrl.slice(i + 1)), (c) => c.charCodeAt(0));
  return new TextDecoder().decode(bytes);
}
