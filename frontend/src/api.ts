// Typed wrappers around the Wails-bound Go backend and runtime events.
// Falls back gracefully in a plain browser (no Wails bridge), so the UI can
// be rendered standalone for design review.

import * as AppBindings from "../bindings/croc-desktop/app";
import { Events, Clipboard } from "@wailsio/runtime";

type WailsApp = {
  PickFiles(): Promise<string[]>;
  PickDirectory(): Promise<string>;
  PathsIsDir(paths: string[]): Promise<string[]>;
  GetDefaultDownloadDir(): Promise<string>;
  StartSend(paths: string[]): Promise<string>;
  StartSendText(text: string): Promise<string>;
  StartReceive(code: string, outDir: string): Promise<void>;
  NormalizeCode(input: string): Promise<string>;
  CancelTransfer(): Promise<void>;
  RespondAccept(accept: boolean): Promise<void>;
  RespondOverwrite(overwrite: boolean): Promise<void>;
  GetQrPng(text: string): Promise<string>;
  GetFileDataURL(path: string): Promise<string>;
  GetHistory(): Promise<HistoryItem[]>;
  ClearHistory(): Promise<void>;
  GetLogs(): Promise<LogEntry[]>;
  ClearLogs(): Promise<void>;
  IsTransferRunning(): Promise<boolean>;
  DecodeCodeFromFile(path: string): Promise<string>;
  DecodeCodeFromBase64(b64: string): Promise<string>;
  PickImage(): Promise<string>;
  GetSettings(): Promise<Settings>;
  SaveSettings(s: Settings): Promise<void>;
  StartRelay(ports: string[], password: string): Promise<void>;
  StopRelay(): Promise<void>;
  GetRelayState(): Promise<RelayState>;
};

// Preview stubs keep the UI renderable for design review when no backend is
// present (plain browser). Live bindings are preferred when the Wails bridge
// is available.
const previewStub: Partial<WailsApp> = {
  GetSettings: () =>
    Promise.resolve({
      relayAddress: "croc.schollz.com",
      relayAddress6: "croc6.schollz.com",
      relayPassword: "pass123",
      curve: "p256",
      hashAlgorithm: "xxhash",
      onlyLocal: false,
      disableLocal: false,
      noCompress: false,
      overwrite: false,
      downloadDir: "",
      socks5: "",
      httpProxy: "",
      zipFolder: false,
      exclude: "",
      throttleUpload: "",
      ip: "",
      theme: "system",
      language: "system",
      savedCodes: [],
    }),
  GetDefaultDownloadDir: () => Promise.resolve(""),
  GetRelayState: () => Promise.resolve({ running: false }),
};

const rejectingProxy = new Proxy(
  {},
  {
    get:
      () =>
      (..._args: unknown[]) =>
        Promise.reject(new Error("backend unavailable")),
  }
);

export const App: WailsApp = new Proxy(previewStub, {
  get: (target, prop: string | symbol) => {
    const bound = (AppBindings as Record<string, unknown>)[prop as string];
    if (typeof bound === "function") {
      return (...args: unknown[]) => {
        const result = (bound as (...a: unknown[]) => Promise<unknown>)(...args);
        if (result && typeof (result as Promise<unknown>).then === "function") {
          return (result as Promise<unknown>).catch((err: unknown) => {
            const stub = (target as Record<string, unknown>)[prop as string];
            if (typeof stub === "function") {
              return (stub as (...a: unknown[]) => unknown)(...args);
            }
            return Promise.reject(err);
          });
        }
        return result;
      };
    }
    return (target as Record<string, unknown>)[prop as string] ?? (rejectingProxy as any)[prop];
  },
}) as unknown as WailsApp;

// E2E bridge: Playwright helpers (and older snippets) call window.go.main.App
// and window.runtime.* — the v2 surface. Production UI uses App / onEvent /
// copyToClipboard above, which talk to v3 bindings and @wailsio/runtime.
// Keep this shim until e2e/ is updated to import bindings directly.
if (typeof window !== "undefined") {
  (window as any).go = { main: { App: AppBindings } };
  (window as any).runtime = {
    EventsOn: (name: string, cb: (data?: any) => void) =>
      Events.On(name, (ev: { data?: any }) => cb(ev?.data)),
    EventsOff: (...names: string[]) => {
      for (const n of names) Events.Off(n);
    },
    EventsEmit: (name: string, data?: any) => Events.Emit(name, data),
    ClipboardSetText: (text: string) =>
      Clipboard.SetText(text)
        .then(() => true)
        .catch(() => false),
  };
}

export interface FileEntry {
  name: string;
  folder: string;
  size: number;
}

export interface AcceptPayload {
  files: FileEntry[];
  totalSize: number;
  totalFolders: number;
  senderId: string;
  isText: boolean;
}

export interface OverwritePayload {
  path: string;
  resumePct: number;
}

export interface ProgressPayload {
  filename: string;
  bytesDone: number;
  bytesTotal: number;
  filesDone: number;
  filesTotal: number;
}

export interface ReceivedFile {
  name: string;
  path: string;
  size: number;
}

export interface DonePayload {
  isText: boolean;
  text?: string;
  files?: ReceivedFile[];
}

export interface HistoryFile {
  name: string;
  path?: string;
  size: number;
}

export interface HistoryItem {
  id: string;
  time: string;
  direction: "send" | "receive";
  status: "completed" | "cancelled" | "error";
  error?: string;
  isText?: boolean;
  text?: string;
  files?: HistoryFile[];
  totalFiles?: number;
  totalSize?: number;
  dir?: string;
}

export interface LogEntry {
  id: string;
  time: string;
  level: "debug" | "info" | "warn" | "error";
  source: string;
  message: string;
}

export interface SavedCode {
  name: string;
  code: string;
}

export interface Settings {
  relayAddress: string;
  relayAddress6: string;
  relayPassword: string;
  curve: string;
  hashAlgorithm: string;
  onlyLocal: boolean;
  disableLocal: boolean;
  noCompress: boolean;
  overwrite: boolean;
  downloadDir: string;
  socks5: string;
  httpProxy: string;
  zipFolder: boolean;
  exclude: string;
  throttleUpload: string;
  ip: string;
  theme: "system" | "light" | "dark";
  language: string;
  savedCodes?: SavedCode[];
}

export interface RelayState {
  running: boolean;
  ports?: string[];
  error?: string;
}

export type TransferStateName = "connecting" | "waiting" | "transferring" | "cancelled";

// Subscribe to a backend event; returns an unsubscribe function.
// v3 delivers a WailsEvent object; we unwrap .data for callers.
export function onEvent<T>(name: string, callback: (data: T) => void): () => void {
  try {
    return Events.On(name, (ev: { data?: T }) => callback(ev?.data as T));
  } catch {
    return () => {};
  }
}

export function copyToClipboard(text: string): Promise<boolean> {
  try {
    return Clipboard.SetText(text)
      .then(() => true)
      .catch(() => navigator.clipboard.writeText(text).then(() => true));
  } catch {
    return navigator.clipboard.writeText(text).then(() => true);
  }
}

export function formatBytes(n: number): string {
  if (!n || n < 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let i = 0;
  let v = n;
  while (v >= 1000 && i < units.length - 1) {
    v /= 1000;
    i++;
  }
  return `${v >= 100 ? Math.round(v) : v.toFixed(1)} ${units[i]}`;
}
