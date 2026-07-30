import { useEffect, useState } from "react";
import { App as Backend } from "./api";
import { useTransfer } from "./useTransfer";
import { applyTheme } from "./theme";
import { setLanguage, useT } from "./i18n";
import SendView from "./views/SendView";
import ReceiveView from "./views/ReceiveView";
import HistoryView from "./views/HistoryView";
import LogsView from "./views/LogsView";
import RelayView from "./views/RelayView";
import SettingsView from "./views/SettingsView";
import { AcceptModal, OverwriteModal } from "./components/Modals";
import Logo from "./components/Logo";

type Tab = "send" | "receive" | "history" | "logs" | "relay" | "settings";

const tabs: { id: Tab; labelKey: string }[] = [
  { id: "send", labelKey: "app.send" },
  { id: "receive", labelKey: "app.receive" },
  { id: "history", labelKey: "app.history" },
  { id: "logs", labelKey: "app.logs" },
  { id: "relay", labelKey: "app.relay" },
  { id: "settings", labelKey: "app.settings" },
];

export default function App() {
  const [tab, setTab] = useState<Tab>("send");
  const transfer = useTransfer();
  const t = useT();

  // apply the persisted theme and language at startup
  useEffect(() => {
    Backend.GetSettings()
      .then((s) => {
        applyTheme(s.theme);
        setLanguage(s.language);
      })
      .catch(() => {});
  }, []);

  // Esc cancels an in-flight transfer from anywhere in the app
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (
        e.key === "Escape" &&
        (transfer.phase === "connecting" ||
          transfer.phase === "waiting" ||
          transfer.phase === "transferring")
      ) {
        Backend.CancelTransfer();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [transfer.phase]);

  return (
    <div className="app">
      {/* N3 side-rail: logo + filled-bar section indicators */}
      <aside className="rail">
        <div className="rail-logo">
          <Logo />
        </div>
        <nav className="rail-nav">
          {tabs.map((item) => (
            <button
              key={item.id}
              className={`rail-item ${tab === item.id ? "rail-item-active" : ""}`}
              onClick={() => setTab(item.id)}
            >
              <span className="rail-indicator" aria-hidden="true" />
              <span className="rail-label">{t(item.labelKey)}</span>
            </button>
          ))}
        </nav>
        <div className="rail-foot">croc v10</div>
      </aside>

      <main className="content">
        {/* views stay mounted (hidden) so their local state survives tab switches */}
        <div className={tab === "send" ? "" : "hidden"}>
          <SendView transfer={transfer} />
        </div>
        <div className={tab === "receive" ? "" : "hidden"}>
          <ReceiveView transfer={transfer} />
        </div>
        <div className={tab === "history" ? "" : "hidden"}>
          <HistoryView />
        </div>
        <div className={tab === "logs" ? "" : "hidden"}>
          <LogsView />
        </div>
        <div className={tab === "relay" ? "" : "hidden"}>
          <RelayView />
        </div>
        <div className={tab === "settings" ? "" : "hidden"}>
          <SettingsView />
        </div>
      </main>

      {/* prompt modals are app-level so they surface on any tab */}
      {transfer.accept && (
        <AcceptModal
          payload={transfer.accept}
          onRespond={(ok) => {
            Backend.RespondAccept(ok);
            transfer.clearAccept();
            if (!ok) transfer.reset();
          }}
        />
      )}
      {transfer.overwrite && (
        <OverwriteModal
          payload={transfer.overwrite}
          onRespond={(ok) => {
            Backend.RespondOverwrite(ok);
            transfer.clearOverwrite();
          }}
        />
      )}
    </div>
  );
}
