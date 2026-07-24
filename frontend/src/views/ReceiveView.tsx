import { useEffect, useRef, useState } from "react";
import { App as Backend, copyToClipboard } from "../api";
import { TransferModel } from "../useTransfer";
import { StatusCard } from "./SendView";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

export default function ReceiveView({
  transfer,
}: {
  transfer: TransferModel;
}) {
  const [code, setCode] = useState("");
  const [outDir, setOutDir] = useState("");
  const [startError, setStartError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);
  const [copied, setCopied] = useState(false);
  const [imgMsg, setImgMsg] = useState<{ kind: "ok" | "err"; text: string } | null>(null);
  const [imgBusy, setImgBusy] = useState(false);
  const viewRef = useRef<HTMLDivElement>(null);

  const active = transfer.direction === "receive" && transfer.phase !== "idle";

  useEffect(() => {
    Backend.GetSettings()
      .then((s) => {
        if (s.downloadDir) return s.downloadDir;
        return Backend.GetDefaultDownloadDir();
      })
      .then(setOutDir)
      .catch(() => {});
  }, []);

  // Paste a screenshot of the sender's QR code anywhere in the app — the
  // code is extracted in the background; the image itself is never shown.
  useEffect(() => {
    const onPaste = (e: ClipboardEvent) => {
      if (!viewRef.current || viewRef.current.offsetParent === null) return;
      const items = e.clipboardData?.items;
      if (!items) return;
      for (const item of items) {
        if (item.kind === "file" && item.type.startsWith("image/")) {
          const file = item.getAsFile();
          if (!file) continue;
          const reader = new FileReader();
          reader.onload = () => decodeImage(String(reader.result));
          reader.readAsDataURL(file);
          e.preventDefault();
          break;
        }
      }
    };
    window.addEventListener("paste", onPaste);
    return () => window.removeEventListener("paste", onPaste);
  }, []);

  const decodeImage = async (data: string) => {
    setImgBusy(true);
    setImgMsg(null);
    try {
      const decoded = await Backend.DecodeCodeFromBase64(data);
      setCode(decoded);
      setImgMsg({ kind: "ok", text: "Code read from image." });
    } catch (e) {
      setImgMsg({ kind: "err", text: errMsg(e) });
    } finally {
      setImgBusy(false);
    }
  };

  const chooseImage = async () => {
    setImgMsg(null);
    try {
      const path = await Backend.PickImage();
      if (!path) return;
      setImgBusy(true);
      try {
        const decoded = await Backend.DecodeCodeFromFile(path);
        setCode(decoded);
        setImgMsg({ kind: "ok", text: "Code read from image." });
      } finally {
        setImgBusy(false);
      }
    } catch (e) {
      setImgMsg({ kind: "err", text: errMsg(e) });
    }
  };

  const browse = async () => {
    const dir = await Backend.PickDirectory();
    if (dir) setOutDir(dir);
  };

  const start = async () => {
    setStartError(null);
    setStarting(true);
    try {
      await Backend.StartReceive(code.trim(), outDir);
      transfer.begin("receive", "");
    } catch (e) {
      setStartError(errMsg(e));
    } finally {
      setStarting(false);
    }
  };

  const copyText = async () => {
    if (transfer.result?.text) {
      await copyToClipboard(transfer.result.text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  };

  if (active) {
    const doneContent =
      transfer.phase === "done" ? (
        transfer.result?.isText ? (
          <div>
            <p className="status status-ok">Text message received:</p>
            <pre className="received-text">{transfer.result.text}</pre>
            <div className="btn-row">
              <button className="btn btn-ghost btn-sm" onClick={copyText}>
                {copied ? "Copied!" : "Copy text"}
              </button>
            </div>
          </div>
        ) : (
          <div>
            <p className="status status-ok">Transfer complete.</p>
            <p className="hint break-all">Saved to {outDir}</p>
          </div>
        )
      ) : undefined;

    return (
      <div className="view">
        <h2 className="view-title">Receive</h2>
        <StatusCard
          transfer={transfer}
          onCancel={Backend.CancelTransfer}
          doneContent={doneContent}
        />
        {(transfer.phase === "done" ||
          transfer.phase === "error" ||
          transfer.phase === "cancelled") && (
          <button className="btn btn-primary" onClick={transfer.reset}>
            Receive another
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="view" ref={viewRef}>
      <h2 className="view-title">Receive</h2>
      <label className="field">
        <span className="field-label">Code phrase</span>
        <input
          className="input code-input"
          placeholder="e.g. 4-word-code-phrase"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && code.trim() && start()}
          autoFocus
        />
      </label>

      <div className="qr-row">
        <span className="hint">QR screenshot? Paste it (Ctrl+V) or</span>
        <button
          className="btn btn-ghost btn-sm"
          onClick={chooseImage}
          disabled={imgBusy}
        >
          {imgBusy ? "Reading…" : "Choose image…"}
        </button>
      </div>
      {imgMsg && (
        <p className={imgMsg.kind === "err" ? "error-text" : "hint qr-hint"}>
          {imgMsg.text}
        </p>
      )}

      <label className="field">
        <span className="field-label">Save to</span>
        <div className="input-row">
          <input
            className="input"
            value={outDir}
            onChange={(e) => setOutDir(e.target.value)}
          />
          <button className="btn btn-ghost" onClick={browse}>
            Browse…
          </button>
        </div>
      </label>

      {startError && <p className="error-text">{startError}</p>}

      <button
        className="btn btn-primary btn-lg"
        disabled={starting || code.trim().length < 6}
        onClick={start}
      >
        {starting ? "Connecting…" : "Receive"}
      </button>
    </div>
  );
}
