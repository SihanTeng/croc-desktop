import { useEffect, useState } from "react";
import { App as Backend, copyToClipboard } from "../api";

export default function CodeDisplay({ code }: { code: string }) {
  const [qr, setQr] = useState<string>("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let alive = true;
    Backend.GetQrPng(code).then((b64) => {
      if (alive) setQr(b64);
    });
    return () => {
      alive = false;
    };
  }, [code]);

  const copy = async () => {
    await copyToClipboard(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="code-display">
      <div className="code-row">
        <code className="code-phrase">{code}</code>
        <button className="btn btn-ghost-dark btn-sm" onClick={copy}>
          {copied ? "Copied ✓" : "Copy"}
        </button>
      </div>
      {qr && (
        <img
          className="code-qr"
          src={`data:image/png;base64,${qr}`}
          alt="QR code of the receive phrase"
        />
      )}
      <p className="hint">The recipient enters this code — or scans the QR — on their device.</p>
    </div>
  );
}
