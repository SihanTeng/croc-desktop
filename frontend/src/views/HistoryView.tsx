import { useCallback, useEffect, useState } from "react";
import { App as Backend, HistoryItem, formatBytes, onEvent } from "../api";
import { useT } from "../i18n";

// file names rendered per entry before collapsing into "…and N more"
const MAX_SHOWN = 5;

export default function HistoryView() {
  const [items, setItems] = useState<HistoryItem[] | null>(null);
  const t = useT();

  const load = useCallback(() => {
    Backend.GetHistory()
      .then((h) => setItems(h ?? []))
      .catch(() => setItems([]));
  }, []);

  useEffect(() => {
    load();
    // refresh when a transfer reaches any terminal state
    const offs = [
      onEvent("transfer:done", load),
      onEvent("transfer:error", load),
      onEvent<string>("transfer:state", (s) => {
        if (s === "cancelled") load();
      }),
    ];
    return () => offs.forEach((off) => off());
  }, [load]);

  const clear = async () => {
    await Backend.ClearHistory();
    load();
  };

  return (
    <div className="view">
      <div className="history-head">
        <h2 className="view-title">{t("history.title")}</h2>
        {items && items.length > 0 && (
          <button className="btn btn-ghost btn-sm" onClick={clear}>
            {t("history.clear")}
          </button>
        )}
      </div>
      {items === null ? (
        <p className="hint">{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p className="hint">{t("history.empty")}</p>
      ) : (
        <ul className="history-list">
          {items.map((item) => (
            <HistoryRow key={item.id} item={item} />
          ))}
        </ul>
      )}
    </div>
  );
}

function HistoryRow({ item }: { item: HistoryItem }) {
  const t = useT();
  const when = new Date(item.time).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
  const extra = (item.totalFiles ?? 0) - Math.min(item.files?.length ?? 0, MAX_SHOWN);
  return (
    <li className="history-item">
      <div className="history-top">
        <span className={`history-arrow history-arrow-${item.direction}`}>
          {item.direction === "send" ? "↑" : "↓"}
        </span>
        <span className="history-verb">
          {item.direction === "send" ? t("history.sent") : t("history.received")}
        </span>
        {item.isText ? (
          <span className="history-size">{t("history.text")}</span>
        ) : (
          item.totalSize != null &&
          item.totalSize > 0 && <span className="history-size">{formatBytes(item.totalSize)}</span>
        )}
        <span className={`history-status history-status-${t(`history.status.${item.status}`)}`}>
          {t(`history.status.${item.status}`)}
        </span>
        <time className="history-time" dateTime={item.time}>
          {when}
        </time>
      </div>
      {item.error && <p className="error-text">{item.error}</p>}
      {item.isText && item.text && <pre className="received-text history-text">{item.text}</pre>}
      {item.files && item.files.length > 0 && (
        <ul className="file-list history-files">
          {item.files.slice(0, MAX_SHOWN).map((f, i) => (
            <li key={i}>
              <span className="file-name">{f.name}</span>
              <span className="file-size">{formatBytes(f.size)}</span>
            </li>
          ))}
          {extra > 0 && <li className="file-more">{t("history.andMore", { n: extra })}</li>}
        </ul>
      )}
      {item.dir && item.status === "completed" && (
        <p className="hint break-all">{t("receive.savedTo", { dir: item.dir })}</p>
      )}
    </li>
  );
}
