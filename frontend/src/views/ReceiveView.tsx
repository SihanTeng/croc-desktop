import { useEffect, useMemo, useRef, useState } from "react";
import { App as Backend, copyToClipboard, formatBytes, ReceivedFile } from "../api";
import { decodeDataUrlText, maxTextPreview, previewKind } from "../filepreview";
import { TransferModel } from "../useTransfer";
import { StatusCard } from "./SendView";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

// ReceivedFileCard shows a received file with a best-effort inline preview
// (image/video/audio/text) and falls back to a plain name+size row.
function ReceivedFileCard({ file }: { file: ReceivedFile }) {
  const kind = previewKind(file.name);
  const [dataUrl, setDataUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!kind) return;
    if (kind === "text" && file.size > maxTextPreview) return;
    let live = true;
    Backend.GetFileDataURL(file.path)
      .then((u) => live && setDataUrl(u))
      .catch(() => live && setFailed(true));
    return () => {
      live = false;
    };
  }, [file.path, file.size, kind]);

  return (
    <div className="recv-file">
      {kind === "image" && dataUrl && <img className="recv-media" src={dataUrl} alt={file.name} />}
      {kind === "video" && dataUrl && <video className="recv-media" src={dataUrl} controls />}
      {kind === "audio" && dataUrl && <audio className="recv-audio" src={dataUrl} controls />}
      {kind === "text" && dataUrl && <TextPreview dataUrl={dataUrl} />}
      <p className="hint break-all recv-name" title={file.path}>
        {file.name} · {formatBytes(file.size)}
        {failed && " (preview unavailable)"}
      </p>
    </div>
  );
}

function TextPreview({ dataUrl }: { dataUrl: string }) {
  const text = useMemo(() => decodeDataUrlText(dataUrl), [dataUrl]);
  return <pre className="received-text">{text}</pre>;
}

export default function ReceiveView({ transfer }: { transfer: TransferModel }) {
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
      // accept anything pasteable: bare code, "croc <code>", "CROC_SECRET=…",
      // or a share link carrying the code
      const normalized = await Backend.NormalizeCode(code);
      if (normalized !== code.trim()) setCode(normalized);
      await Backend.StartReceive(normalized, outDir);
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
            {transfer.result?.files && transfer.result.files.length > 0 && (
              <div className="recv-files">
                {transfer.result.files.map((f) => (
                  <ReceivedFileCard key={f.path} file={f} />
                ))}
              </div>
            )}
          </div>
        )
      ) : undefined;

    return (
      <div className="view">
        <h2 className="view-title">Receive</h2>
        <StatusCard
          transfer={transfer}
          onCancel={() => Backend.CancelTransfer()}
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
          placeholder="Code, croc command, or link"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && code.trim() && start()}
          autoFocus
        />
      </label>

      <div className="qr-row">
        <span className="hint">QR screenshot? Paste it (Ctrl+V) or</span>
        <button className="btn btn-ghost btn-sm" onClick={chooseImage} disabled={imgBusy}>
          {imgBusy ? "Reading…" : "Choose image…"}
        </button>
      </div>
      {imgMsg && (
        <p className={imgMsg.kind === "err" ? "error-text" : "hint qr-hint"}>{imgMsg.text}</p>
      )}

      <label className="field">
        <span className="field-label">Save to</span>
        <div className="input-row">
          <input className="input" value={outDir} onChange={(e) => setOutDir(e.target.value)} />
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
