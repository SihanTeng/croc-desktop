// Typed wrappers around the Wails-bound Go backend and runtime events.
// Falls back gracefully in a plain browser (no window.go / window.runtime),
// so the UI can be rendered standalone for design review.

type WailsApp = Window["go"]["main"]["App"];

// The backend binding object. Always dereferenced at call time, so the UI
// keeps working across wails-dev app restarts (which silently replace
// window.go while the page stays open). In a plain browser (no backend),
// preview stubs keep the UI renderable for design review.
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
  get: (target, prop) => {
    const live = window.go?.main?.App;
    if (live) return live[prop as keyof WailsApp];
    return (target as any)[prop] ?? (rejectingProxy as any)[prop];
  },
}) as unknown as WailsApp;

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
export function onEvent<T>(name: string, callback: (data: T) => void): () => void {
  if (!window.runtime?.EventsOn) return () => {};
  return window.runtime.EventsOn(name, (data?: any) => callback(data as T));
}

export function copyToClipboard(text: string): Promise<boolean> {
  if (window.runtime?.ClipboardSetText) {
    return window.runtime.ClipboardSetText(text);
  }
  return navigator.clipboard.writeText(text).then(() => true);
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
