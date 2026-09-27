import { useEffect, useState } from "react";
import { App as Backend, copyToClipboard } from "../api";
import { useT } from "../i18n";

export default function CodeDisplay({ code }: { code: string }) {
  const t = useT();
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

  // Match the local OS. Unix v11 CLI requires CROC_SECRET instead of a
  // positional code; Windows still accepts positional codes.
  const copyCommand = async () => {
    const quoted = "'" + code.replace(/'/g, "'\\''") + "'";
    const command = /Windows/i.test(navigator.userAgent)
      ? `croc ${code}`
      : `CROC_SECRET=${quoted} croc`;
    await copyToClipboard(command);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 1500);
  };

  return (
    <div className="code-display">
      <div className="code-row">
        <code className="code-phrase">{code}</code>
        <div className="btn-row">
          <button className="btn btn-ghost-dark btn-sm" onClick={copy}>
            {copied ? t("code.copied") : t("code.copy")}
          </button>
          <button className="btn btn-ghost-dark btn-sm" onClick={copyCommand}>
            {copiedCmd ? t("code.copied") : t("code.copyCommand")}
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
      <p className="hint">{t("code.hint")}</p>
    </div>
  );
}
