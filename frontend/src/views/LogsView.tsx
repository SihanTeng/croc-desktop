import { useCallback, useEffect, useRef, useState } from "react";
import { App as Backend, LogEntry, onEvent } from "../api";
import { useT } from "../i18n";

// levelRank orders the levels; the filter keeps entries at or above it
const levelRank: Record<LogEntry["level"], number> = { debug: 0, info: 1, warn: 2, error: 3 };

const filters = [
  { id: 0, labelKey: "logs.all" },
  { id: 1, labelKey: "logs.infoPlus" },
  { id: 2, labelKey: "logs.warnPlus" },
  { id: 3, labelKey: "logs.error" },
] as const;

// keep the client-side mirror the same size as the backend buffer
const MAX_ENTRIES = 500;

export default function LogsView() {
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const t = useT();
  const [minLevel, setMinLevel] = useState<number>(0);
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    Backend.GetLogs()
      .then((l) => setEntries(l ?? []))
      .catch(() => {});
    return onEvent<LogEntry>("log:entry", (entry) =>
      setEntries((prev) => [...prev, entry].slice(-MAX_ENTRIES))
    );
  }, []);

  // follow the tail as new entries stream in
  useEffect(() => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [entries, minLevel]);

  const clear = useCallback(async () => {
    await Backend.ClearLogs();
    setEntries([]);
  }, []);

  const shown = entries.filter((e) => levelRank[e.level] >= minLevel);

  return (
    <div className="view">
      <div className="logs-head">
        <h2 className="view-title">{t("logs.title")}</h2>
        <div className="logs-controls">
          <div className="segmented">
            {filters.map((f) => (
              <button
                key={f.id}
                className={minLevel === f.id ? "segment segment-active" : "segment"}
                onClick={() => setMinLevel(f.id)}
              >
                {t(f.labelKey)}
              </button>
            ))}
          </div>
          {entries.length > 0 && (
            <button className="btn btn-ghost btn-sm" onClick={clear}>
              Clear logs
            </button>
          )}
        </div>
      </div>
      {shown.length === 0 ? (
        <p className="hint">{entries.length === 0 ? t("logs.empty") : t("logs.nothingAtLevel")}</p>
      ) : (
        <div className="logs-list" ref={listRef}>
          {shown.map((e) => (
            <div className="log-row" key={e.id}>
              <span className="log-time">
                {new Date(e.time).toLocaleTimeString(undefined, { hour12: false })}
              </span>
              <span className={`log-level log-level-${e.level}`}>{e.level}</span>
              <span className="log-source">{e.source}</span>
              <span className="log-message">{e.message}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
