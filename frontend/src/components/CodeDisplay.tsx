import { useEffect, useState } from "react";
import { App as Backend, copyToClipboard } from "../api";

export default function CodeDisplay({ code }: { code: string }) {
  const [qr, setQr] = useState<string>("");
  const [copied, setCopied] = useState(false);
  const [copiedCmd, setCopiedCmd] = useState(false);

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

  // for recipients on the croc CLI
  const copyCommand = async () => {
    await copyToClipboard(`croc ${code}`);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 1500);
  };

  return (
    <div className="code-display">
      <div className="code-row">
        <code className="code-phrase">{code}</code>
        <div className="btn-row">
          <button className="btn btn-ghost-dark btn-sm" onClick={copy}>
            {copied ? "Copied ✓" : "Copy"}
          </button>
          <button className="btn btn-ghost-dark btn-sm" onClick={copyCommand}>
            {copiedCmd ? "Copied ✓" : "Copy command"}
          </button>
        </div>
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
