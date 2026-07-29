import { useEffect, useState } from "react";
import { App as Backend, Settings } from "../api";

const CURVES = ["p256", "p384", "p521", "siec", "ed25519"];
const HASHES = ["xxhash", "imohash", "md5", "highway"];

export default function SettingsView() {
  const [s, setS] = useState<Settings | null>(null);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Backend.GetSettings().then(setS);
  }, []);

  if (!s) return <div className="view" />;

  const update = <K extends keyof Settings>(key: K, value: Settings[K]) => {
    setS({ ...s, [key]: value });
    setSaved(false);
  };

  const browseDir = async () => {
    const dir = await Backend.PickDirectory();
    if (dir) update("downloadDir", dir);
  };

  const save = async () => {
    setError(null);
    try {
      await Backend.SaveSettings(s);
      setSaved(true);
      setTimeout(() => setSaved(false), 1500);
    } catch (e) {
      setError(String(e));
    }
  };

  return (
    <div className="view">
      <h2 className="view-title">Settings</h2>
      <div className="grid-2">
        <label className="field">
          <span className="field-label">Relay address (IPv4)</span>
          <input
            className="input"
            value={s.relayAddress}
            onChange={(e) => update("relayAddress", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">Relay address (IPv6)</span>
          <input
            className="input"
            value={s.relayAddress6}
            onChange={(e) => update("relayAddress6", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">Relay password</span>
          <input
            className="input"
            type="password"
            value={s.relayPassword}
            onChange={(e) => update("relayPassword", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">Default download folder</span>
          <div className="input-row">
            <input
              className="input"
              value={s.downloadDir}
              onChange={(e) => update("downloadDir", e.target.value)}
            />
            <button className="btn btn-ghost" onClick={browseDir}>
              Browse…
            </button>
          </div>
        </label>
        <label className="field">
          <span className="field-label">SOCKS5 proxy</span>
          <input
            className="input"
            placeholder="user:pass@host:port (optional)"
            value={s.socks5}
            onChange={(e) => update("socks5", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">HTTP proxy</span>
          <input
            className="input"
            placeholder="user:pass@host:port (optional)"
            value={s.httpProxy}
            onChange={(e) => update("httpProxy", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">Encryption curve</span>
          <select
            className="input"
            value={s.curve}
            onChange={(e) => update("curve", e.target.value)}
          >
            {CURVES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span className="field-label">Hash algorithm</span>
          <select
            className="input"
            value={s.hashAlgorithm}
            onChange={(e) => update("hashAlgorithm", e.target.value)}
          >
            {HASHES.map((h) => (
              <option key={h} value={h}>
                {h}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="checks">
        {(
          [
            ["onlyLocal", "Local network only"],
            ["disableLocal", "Disable local relay"],
            ["noCompress", "Disable compression"],
            ["overwrite", "Overwrite without asking"],
          ] as [keyof Settings, string][]
        ).map(([key, label]) => (
          <label className="check" key={key}>
            <input
              type="checkbox"
              checked={Boolean(s[key])}
              onChange={(e) => update(key, e.target.checked as never)}
            />
            <span>{label}</span>
          </label>
        ))}
      </div>

      {error && <p className="error-text">{error}</p>}

      <button className="btn btn-primary btn-lg" onClick={save}>
        {saved ? "Saved!" : "Save settings"}
      </button>
    </div>
  );
}
