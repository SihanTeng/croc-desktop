import { useEffect, useState } from "react";
import { App as Backend, RelayState, onEvent } from "../api";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

export default function RelayView() {
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
      <h2 className="view-title">Relay</h2>
      <div className="card">
        <div className="relay-status">
          <span className={`dot ${state.running ? "dot-on" : "dot-off"}`} />
          <span className="status">{state.running ? "Relay running" : "Relay stopped"}</span>
          {state.running && <span className="relay-ports">{state.ports?.join(", ")}</span>}
        </div>
        {state.running && (
          <p className="hint">
            Others can use it as relay address <code>&lt;this-machine-ip&gt;:{firstPort}</code>
          </p>
        )}
      </div>

      <label className="field">
        <span className="field-label">Ports (comma separated, first is control)</span>
        <input
          className="input"
          value={ports}
          disabled={state.running}
          onChange={(e) => setPorts(e.target.value)}
        />
      </label>

      <label className="field">
        <span className="field-label">Password (blank uses the croc default)</span>
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
        {state.running ? "Stop relay" : "Start relay"}
      </button>
    </div>
  );
}
