import { useEffect, useState } from "react";
import { App as Backend, RelayState, onEvent } from "../api";
import { useT } from "../i18n";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

export default function RelayView() {
  const t = useT();
  const [ports, setPorts] = useState("9009,9010,9011,9012,9013");
  const [password, setPassword] = useState("");
  const [state, setState] = useState<RelayState>({ running: false });
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Backend.GetRelayState()
      .then(setState)
      .catch(() => {});
    return onEvent<RelayState>("relay:state", (s) => {
      setState(s);
      if (s.error) setError(s.error);
    });
  }, []);

  const toggle = async () => {
    setError(null);
    try {
      if (state.running) {
        await Backend.StopRelay();
        setState({ running: false });
      } else {
        const list = ports
          .split(",")
          .map((p) => p.trim())
          .filter(Boolean);
        await Backend.StartRelay(list, password);
      }
    } catch (e) {
      setError(errMsg(e));
    }
  };

  const firstPort = (state.ports?.[0] ?? ports.split(",")[0] ?? "").trim();

  return (
    <div className="view">
      <h2 className="view-title">{t("relay.title")}</h2>
      <div className="card">
        <div className="relay-status">
          <span className={`dot ${state.running ? "dot-on" : "dot-off"}`} />
          <span className="status">{state.running ? t("relay.running") : t("relay.stopped")}</span>
          {state.running && <span className="relay-ports">{state.ports?.join(", ")}</span>}
        </div>
        {state.running && (
          <p className="hint">
            {t("relay.hint")} <code>&lt;this-machine-ip&gt;:{firstPort}</code>
          </p>
        )}
      </div>

      <label className="field">
        <span className="field-label">{t("relay.portsLabel")}</span>
        <input
          className="input"
          value={ports}
          disabled={state.running}
          onChange={(e) => setPorts(e.target.value)}
        />
      </label>

      <label className="field">
        <span className="field-label">{t("relay.passwordLabel")}</span>
        <input
          className="input"
          type="password"
          value={password}
          disabled={state.running}
          onChange={(e) => setPassword(e.target.value)}
        />
      </label>

      {error && <p className="error-text">{error}</p>}

      <button
        className={`btn btn-lg ${state.running ? "btn-danger" : "btn-primary"}`}
        onClick={toggle}
      >
        {state.running ? t("relay.stop") : t("relay.start")}
      </button>
    </div>
  );
}
