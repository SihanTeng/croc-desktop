// Type declarations for the APIs Wails injects into the window object.
// The Go backend methods bound in gui/app.go appear under window.go.main.App;
// the Wails runtime appears under window.runtime.

export {};

declare global {
  interface Window {
    go: {
      main: {
        App: {
          PickFiles(): Promise<string[]>;
          PickDirectory(): Promise<string>;
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
          StartRelay(ports: string[], password: string): Promise<void>;
          StopRelay(): Promise<void>;
          GetRelayState(): Promise<import("./api").RelayState>;
        };
      };
    };
    runtime: {
      EventsOn(eventName: string, callback: (data?: any) => void): () => void;
      EventsOff(eventName: string, ...additionalEventNames: string[]): void;
      ClipboardSetText(text: string): Promise<boolean>;
    };
  }
}
