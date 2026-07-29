import { useEffect, useState } from "react";
import { App as Backend, onEvent } from "../api";
import { TransferModel } from "../useTransfer";
import DropZone from "../components/DropZone";
import CodeDisplay from "../components/CodeDisplay";
import ProgressBar from "../components/ProgressBar";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

export default function SendView({ transfer }: { transfer: TransferModel }) {
  const [mode, setMode] = useState<"files" | "text">("files");
  const [paths, setPaths] = useState<string[]>([]);
  const [text, setText] = useState("");
  const [startError, setStartError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);

  const active = transfer.direction === "send" && transfer.phase !== "idle";

  // native file drops arrive as paths from the backend
  useEffect(() => {
    return onEvent<string[]>("files:dropped", (dropped) => {
      setPaths((p) => Array.from(new Set([...p, ...dropped])));
    });
  }, []);

  const addFiles = async () => {
    const picked = await Backend.PickFiles();
    if (picked?.length) setPaths((p) => Array.from(new Set([...p, ...picked])));
  };
  const addFolder = async () => {
    const dir = await Backend.PickDirectory();
    if (dir) setPaths((p) => (p.includes(dir) ? p : [...p, dir]));
  };
  const removePath = (path: string) => setPaths((p) => p.filter((x) => x !== path));

  const start = async () => {
    setStartError(null);
    setStarting(true);
    try {
      const code =
        mode === "files" ? await Backend.StartSend(paths) : await Backend.StartSendText(text);
      transfer.begin("send", code);
    } catch (e) {
      setStartError(errMsg(e));
    } finally {
      setStarting(false);
    }
  };

  const startOver = () => {
    transfer.reset();
    setPaths([]);
    setText("");
    setStartError(null);
  };

  if (active) {
    return (
      <div className="view">
        <h2 className="view-title">Send</h2>
        <CodeDisplay code={transfer.code} />
        <StatusCard transfer={transfer} onCancel={Backend.CancelTransfer} />
        {(transfer.phase === "done" ||
          transfer.phase === "error" ||
          transfer.phase === "cancelled") && (
          <div className="btn-row">
            <button className="btn btn-primary" onClick={startOver}>
              Send another
            </button>
            {transfer.phase === "done" && (
              <button className="btn btn-ghost" disabled={starting} onClick={start}>
                Send same {mode === "files" ? "files" : "text"} again
              </button>
            )}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="view">
      <h2 className="view-title">Send</h2>
      <div className="segmented">
        <button
          className={mode === "files" ? "segment segment-active" : "segment"}
          onClick={() => setMode("files")}
        >
          Files
        </button>
        <button
          className={mode === "text" ? "segment segment-active" : "segment"}
          onClick={() => setMode("text")}
        >
          Text
        </button>
      </div>

      {mode === "files" ? (
        <>
          <DropZone>
            <div className="dropzone-inner">
              <svg
                className="dropzone-glyph"
                width="30"
                height="30"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M12 3v12" />
                <path d="m7 10 5 5 5-5" />
                <path d="M4 19h16" />
              </svg>
              <p>Drag &amp; drop files or folders here</p>
              <p className="hint">or</p>
              <div className="btn-row">
                <button className="btn btn-ghost" onClick={addFiles}>
                  Choose files
                </button>
                <button className="btn btn-ghost" onClick={addFolder}>
                  Choose folder
                </button>
              </div>
            </div>
          </DropZone>
          {paths.length > 0 && (
            <ul className="file-list">
              {paths.map((p) => (
                <li key={p}>
                  <span className="file-name break-all">{p}</span>
                  <button className="btn btn-ghost btn-sm" onClick={() => removePath(p)}>
                    ✕
                  </button>
                </li>
              ))}
            </ul>
          )}
        </>
      ) : (
        <textarea
          className="text-input"
          placeholder="Type the text to send…"
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && (e.ctrlKey || e.metaKey) && text.trim() && start()}
          rows={6}
        />
      )}

      {startError && <p className="error-text">{startError}</p>}

      <button
        className="btn btn-primary btn-lg"
        disabled={starting || (mode === "files" ? paths.length === 0 : text.trim() === "")}
        onClick={start}
      >
        {starting ? "Starting…" : mode === "files" ? "Send files" : "Send text"}
      </button>
    </div>
  );
}

// StatusCard renders the state of an in-flight transfer (used by Send and
// Receive views).
export function StatusCard({
  transfer,
  onCancel,
  doneContent,
}: {
  transfer: TransferModel;
  onCancel: () => void;
  doneContent?: React.ReactNode;
}) {
  const cancellable =
    transfer.phase === "connecting" ||
    transfer.phase === "waiting" ||
    transfer.phase === "transferring";

  // re-evaluate the stall hint periodically; the interval callback (not
  // render) is where Date.now() may be read
  const [stalled, setStalled] = useState(false);
  useEffect(() => {
    if (transfer.phase !== "transferring" || transfer.lastProgressAt === null) return;
    const at = transfer.lastProgressAt;
    const t = setInterval(() => setStalled(Date.now() - at > 15000), 5000);
    return () => clearInterval(t);
  }, [transfer.phase, transfer.lastProgressAt]);

  const verifying =
    transfer.phase === "transferring" &&
    transfer.progress !== null &&
    transfer.progress.bytesTotal > 0 &&
    transfer.progress.bytesDone >= transfer.progress.bytesTotal;

  return (
    <div className="card">
      {transfer.phase === "connecting" && <p className="status pulse">Connecting…</p>}
      {transfer.phase === "waiting" && (
        <p className="status pulse">
          {transfer.direction === "send"
            ? "Waiting for recipient — share the code"
            : "Waiting for the sender…"}
        </p>
      )}
      {transfer.phase === "transferring" &&
        (transfer.progress ? (
          <>
            <ProgressBar progress={transfer.progress} />
            {verifying && <p className="status pulse">Verifying…</p>}
          </>
        ) : (
          <p className="status pulse">Transferring…</p>
        ))}
      {transfer.phase === "transferring" && stalled && (
        <p className="hint">
          {verifying
            ? "Still working — verifying data…"
            : "Still working — large files can take a moment."}
        </p>
      )}
      {transfer.phase === "done" &&
        (doneContent ?? <p className="status status-ok">Transfer complete.</p>)}
      {transfer.phase === "error" && <p className="status status-err">{transfer.error}</p>}
      {transfer.phase === "cancelled" && <p className="status">Transfer cancelled.</p>}
      {cancellable && (
        <div className="btn-row">
          <button className="btn btn-ghost" onClick={onCancel}>
            Cancel
          </button>
        </div>
      )}
    </div>
  );
}
