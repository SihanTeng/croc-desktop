import { useState } from "react";
import { App as Backend } from "./api";
import { useTransfer } from "./useTransfer";
import SendView from "./views/SendView";
import ReceiveView from "./views/ReceiveView";
import HistoryView from "./views/HistoryView";
import RelayView from "./views/RelayView";
import SettingsView from "./views/SettingsView";
import { AcceptModal, OverwriteModal } from "./components/Modals";
import Logo from "./components/Logo";

type Tab = "send" | "receive" | "history" | "relay" | "settings";

const tabs: { id: Tab; label: string }[] = [
  { id: "send", label: "Send" },
  { id: "receive", label: "Receive" },
  { id: "history", label: "History" },
  { id: "relay", label: "Relay" },
  { id: "settings", label: "Settings" },
];

export default function App() {
  const [tab, setTab] = useState<Tab>("send");
  const transfer = useTransfer();

  return (
    <div className="app">
      {/* N3 side-rail: logo + filled-bar section indicators */}
      <aside className="rail">
        <div className="rail-logo">
          <Logo />
        </div>
        <nav className="rail-nav">
          {tabs.map((t) => (
            <button
              key={t.id}
              className={`rail-item ${tab === t.id ? "rail-item-active" : ""}`}
              onClick={() => setTab(t.id)}
            >
              <span className="rail-indicator" aria-hidden="true" />
              <span className="rail-label">{t.label}</span>
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
