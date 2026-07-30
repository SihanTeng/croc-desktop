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

const tabs: { id: Tab; labelKey: string; icon: string }[] = [
  { id: "send", labelKey: "app.send", icon: "send" },
  { id: "receive", labelKey: "app.receive", icon: "receive" },
  { id: "history", labelKey: "app.history", icon: "history" },
  { id: "logs", labelKey: "app.logs", icon: "logs" },
  { id: "relay", labelKey: "app.relay", icon: "relay" },
  { id: "settings", labelKey: "app.settings", icon: "settings" },
];

function TabIcon({ name }: { name: string }) {
  // Compact monoline glyphs for the mobile bottom bar (24×24 viewBox).
  const common = {
    width: 22,
    height: 22,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.75,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
  };
  switch (name) {
    case "send":
      return (
        <svg {...common}>
          <path d="M12 19V5" />
          <path d="M5 12l7-7 7 7" />
        </svg>
      );
    case "receive":
      return (
        <svg {...common}>
          <path d="M12 5v14" />
          <path d="M19 12l-7 7-7-7" />
        </svg>
      );
    case "history":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="8" />
          <path d="M12 8v4l3 2" />
        </svg>
      );
    case "logs":
      return (
        <svg {...common}>
          <path d="M8 6h12" />
          <path d="M8 12h12" />
          <path d="M8 18h12" />
          <path d="M4 6h.01" />
          <path d="M4 12h.01" />
          <path d="M4 18h.01" />
        </svg>
      );
    case "relay":
      return (
        <svg {...common}>
          <circle cx="6" cy="12" r="2.5" />
          <circle cx="18" cy="6" r="2.5" />
          <circle cx="18" cy="18" r="2.5" />
          <path d="M8.2 11l7.3-4" />
          <path d="M8.2 13l7.3 4" />
        </svg>
      );
    case "settings":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="3" />
          <path d="M12 3v2.2M12 18.8V21M4.9 6.5l1.6 1.6M17.5 15.9l1.6 1.6M3 12h2.2M18.8 12H21M4.9 17.5l1.6-1.6M17.5 8.1l1.6-1.6" />
        </svg>
      );
    default:
      return null;
  }
}

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

  const navButtons = tabs.map((item) => (
    <button
      key={item.id}
      type="button"
      className={`rail-item ${tab === item.id ? "rail-item-active" : ""}`}
      onClick={() => setTab(item.id)}
      aria-current={tab === item.id ? "page" : undefined}
    >
      <span className="rail-indicator" aria-hidden="true" />
      <span className="nav-icon" aria-hidden="true">
        <TabIcon name={item.icon} />
      </span>
      <span className="rail-label">{t(item.labelKey)}</span>
    </button>
  ));

  return (
    <div className="app">
      {/* Desktop: N3 side-rail. Mobile: hidden — bottom-nav takes over. */}
      <aside className="rail" aria-label={t("app.nav")}>
        <div className="rail-logo">
          <Logo />
        </div>
        <nav className="rail-nav">{navButtons}</nav>
        <div className="rail-foot">croc v10</div>
      </aside>

      <main className="content">
        <header className="mobile-top">
          <Logo />
        </header>
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

      {/* Mobile bottom tab bar (phone / narrow tablet / Wails mobile). */}
      <nav className="bottom-nav" aria-label={t("app.nav")}>
        {tabs.map((item) => (
          <button
            key={item.id}
            type="button"
            className={`bottom-nav-item ${tab === item.id ? "bottom-nav-item-active" : ""}`}
            onClick={() => setTab(item.id)}
            aria-current={tab === item.id ? "page" : undefined}
          >
            <span className="nav-icon" aria-hidden="true">
              <TabIcon name={item.icon} />
            </span>
            <span className="bottom-nav-label">{t(item.labelKey)}</span>
          </button>
        ))}
      </nav>

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
