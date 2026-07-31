// Type declarations for Wails v3 globals used by E2E helpers and the optional
// window.go bridge installed in api.ts. Production code imports bindings and
// @wailsio/runtime directly.

export {};

declare global {
  interface Window {
    go: {
      main: {
        App: {
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
          GetHistory(): Promise<import("./api").HistoryItem[]>;
          ClearHistory(): Promise<void>;
          GetLogs(): Promise<import("./api").LogEntry[]>;
          ClearLogs(): Promise<void>;
          IsTransferRunning(): Promise<boolean>;
          DecodeCodeFromFile(path: string): Promise<string>;
          DecodeCodeFromBase64(b64: string): Promise<string>;
          PickImage(): Promise<string>;
          GetSettings(): Promise<import("./api").Settings>;
          SaveSettings(s: import("./api").Settings): Promise<void>;
          GetAppInfo(): Promise<import("./api").AppInfo>;
          StartRelay(ports: string[], password: string): Promise<void>;
          StopRelay(): Promise<void>;
          GetRelayState(): Promise<import("./api").RelayState>;
        };
      };
    };
    // Optional v2-compat surface for E2E helpers that emit runtime events.
    runtime?: {
      EventsOn(eventName: string, callback: (data?: any) => void): () => void;
      EventsOff(eventName: string, ...additionalEventNames: string[]): void;
      EventsEmit?(eventName: string, data?: any): Promise<boolean> | void;
      ClipboardSetText(text: string): Promise<boolean>;
    };
  }
}
