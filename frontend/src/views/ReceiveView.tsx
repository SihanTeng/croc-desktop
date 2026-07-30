import { useCallback, useEffect, useRef, useState } from "react";
import { App as Backend, copyToClipboard, Settings } from "../api";
import { TransferModel } from "../useTransfer";
import { useT } from "../i18n";
import { FileRow } from "../components/FileRow";
import CameraScan from "../components/CameraScan";
import { StatusCard } from "./SendView";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

export default function ReceiveView({ transfer }: { transfer: TransferModel }) {
  const [code, setCode] = useState("");
  const [outDir, setOutDir] = useState("");
  const [startError, setStartError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);
  const [copied, setCopied] = useState(false);
  const [imgMsg, setImgMsg] = useState<{ kind: "ok" | "err"; text: string } | null>(null);
  const [imgBusy, setImgBusy] = useState(false);
  const [cameraOpen, setCameraOpen] = useState(false);
  const [settings, setSettings] = useState<Settings | null>(null);
  const viewRef = useRef<HTMLDivElement>(null);
  const t = useT();

  const active = transfer.direction === "receive" && transfer.phase !== "idle";

  useEffect(() => {
    Backend.GetSettings()
      .then((s) => {
        setSettings(s);
        if (s.downloadDir) return s.downloadDir;
        return Backend.GetDefaultDownloadDir();
      })
      .then(setOutDir)
      .catch(() => {});
  }, []);

  const savedCodes = settings?.savedCodes ?? [];

  const saveCode = async () => {
    if (!settings) return;
    const c = code.trim();
    if (c.length < 6 || savedCodes.some((x) => x.code === c)) return;
    const next = { ...settings, savedCodes: [...savedCodes, { name: c, code: c }] };
    try {
      await Backend.SaveSettings(next);
      setSettings(next);
    } catch {
      /* settings validation errors surface in Settings view */
    }
  };

  const removeCode = async (c: string) => {
    if (!settings) return;
    const next = { ...settings, savedCodes: savedCodes.filter((x) => x.code !== c) };
    try {
      await Backend.SaveSettings(next);
      setSettings(next);
    } catch {
      /* ignore */
    }
  };

  const decodeImage = useCallback(
    async (data: string) => {
      setImgBusy(true);
      setImgMsg(null);
      try {
        const decoded = await Backend.DecodeCodeFromBase64(data);
        setCode(decoded);
        setImgMsg({ kind: "ok", text: t("receive.codeRead") });
      } catch (e) {
        setImgMsg({ kind: "err", text: errMsg(e) });
      } finally {
        setImgBusy(false);
      }
    },
    [t]
  );

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
  }, [decodeImage]);

  const chooseImage = async () => {
    setImgMsg(null);
    try {
      const path = await Backend.PickImage();
      if (!path) return;
      setImgBusy(true);
      try {
        const decoded = await Backend.DecodeCodeFromFile(path);
        setCode(decoded);
        setImgMsg({ kind: "ok", text: t("receive.codeRead") });
      } finally {
        setImgBusy(false);
      }
    } catch (e) {
      setImgMsg({ kind: "err", text: errMsg(e) });
    }
  };

  const onCameraCode = async (raw: string) => {
    setCameraOpen(false);
    setImgMsg(null);
    try {
      // QR text may be a bare code or a full "croc …" / link — normalize
      const normalized = await Backend.NormalizeCode(raw);
      setCode(normalized);
      setImgMsg({ kind: "ok", text: t("receive.codeRead") });
    } catch {
      setCode(raw.trim());
      setImgMsg({ kind: "ok", text: t("receive.codeRead") });
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
            <p className="status status-ok">{t("receive.textReceived")}</p>
            <pre className="received-text">{transfer.result.text}</pre>
            <div className="btn-row">
              <button className="btn btn-ghost btn-sm" onClick={copyText}>
                {copied ? t("common.copied") : t("receive.copyText")}
              </button>
            </div>
          </div>
        ) : (
          <div>
            <p className="status status-ok">{t("status.complete")}</p>
            <p className="hint break-all">{t("receive.savedTo", { dir: outDir })}</p>
            {transfer.result?.files && transfer.result.files.length > 0 && (
              <div className="recv-files file-rows">
                {transfer.result.files.map((f) => (
                  <FileRow key={f.path} name={f.name} path={f.path} size={f.size} defaultOpen />
                ))}
              </div>
            )}
          </div>
        )
      ) : undefined;

    return (
      <div className="view">
        <h2 className="view-title">{t("receive.title")}</h2>
        <StatusCard
          transfer={transfer}
          onCancel={() => Backend.CancelTransfer()}
          doneContent={doneContent}
        />
        {(transfer.phase === "done" ||
          transfer.phase === "error" ||
          transfer.phase === "cancelled") && (
          <button className="btn btn-primary" onClick={transfer.reset}>
            {t("receive.another")}
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="view" ref={viewRef}>
      <h2 className="view-title">{t("receive.title")}</h2>
      <label className="field">
        <span className="field-label">{t("receive.codeLabel")}</span>
        <input
          className="input code-input"
          placeholder={t("receive.codePlaceholder")}
          value={code}
          onChange={(e) => setCode(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && code.trim() && start()}
          autoFocus
        />
      </label>

      <div className="saved-row">
        <button
          className="btn btn-ghost btn-sm"
          disabled={code.trim().length < 6 || savedCodes.some((x) => x.code === code.trim())}
          onClick={saveCode}
          title={t("receive.saveCodeTitle")}
        >
          {t("receive.saveCode")}
        </button>
        {savedCodes.length > 0 && (
          <div className="chips">
            {savedCodes.map((sc) => (
              <span className="chip" key={sc.code}>
                <button className="chip-pick" title={sc.code} onClick={() => setCode(sc.code)}>
                  {sc.name}
                </button>
                <button
                  className="chip-x"
                  onClick={() => removeCode(sc.code)}
                  aria-label={t("receive.removeCode", { name: sc.name })}
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="qr-row">
        <span className="hint">{t("receive.qrHint")}</span>
        <div className="btn-row qr-actions">
          <button
            type="button"
            className="btn btn-ghost btn-sm"
            onClick={() => {
              setImgMsg(null);
              setCameraOpen(true);
            }}
            disabled={imgBusy}
          >
            {t("receive.scanCamera")}
          </button>
          <button
            type="button"
            className="btn btn-ghost btn-sm"
            onClick={chooseImage}
            disabled={imgBusy}
          >
            {imgBusy ? t("receive.reading") : t("receive.chooseImage")}
          </button>
        </div>
      </div>
      {imgMsg && (
        <p className={imgMsg.kind === "err" ? "error-text" : "hint qr-hint"}>{imgMsg.text}</p>
      )}

      {cameraOpen && <CameraScan onCode={onCameraCode} onClose={() => setCameraOpen(false)} />}

      <label className="field">
        <span className="field-label">{t("receive.saveTo")}</span>
        <div className="input-row">
          <input className="input" value={outDir} onChange={(e) => setOutDir(e.target.value)} />
          <button className="btn btn-ghost" onClick={browse}>
            {t("common.browse")}
          </button>
        </div>
      </label>

      {startError && <p className="error-text">{startError}</p>}

      <button
        className="btn btn-primary btn-lg"
        disabled={starting || code.trim().length < 6}
        onClick={start}
      >
        {starting ? t("receive.connecting") : t("receive.receive")}
      </button>
    </div>
  );
}
