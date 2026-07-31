import { useEffect, useState } from "react";
import { App as Backend, AppInfo, Settings } from "../api";
import { applyTheme } from "../theme";
import { availableLanguages, setLanguage, useT } from "../i18n";

const CURVES = ["p256", "p384", "p521", "siec", "ed25519"];
const HASHES = ["xxhash", "imohash", "md5", "highway"];

// display names for the language select, keyed by locale file name
const languageNames: Record<string, string> = {
  en: "English",
  "zh-CN": "简体中文",
  "zh-TW": "繁體中文",
  es: "Español",
  fr: "Français",
  de: "Deutsch",
  ja: "日本語",
};

export default function SettingsView() {
  const t = useT();
  const [s, setS] = useState<Settings | null>(null);
  const [info, setInfo] = useState<AppInfo | null>(null);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Backend.GetSettings().then(setS);
    Backend.GetAppInfo()
      .then(setInfo)
      .catch(() => setInfo(null));
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
      applyTheme(s.theme);
      setLanguage(s.language);
      setSaved(true);
      setTimeout(() => setSaved(false), 1500);
    } catch (e) {
      setError(String(e));
    }
  };

  const relaySummary =
    [s.relayAddress, s.relayAddress6].filter((x) => x && x.trim()).join(" · ") ||
    "—";

  return (
    <div className="view">
      <h2 className="view-title">{t("settings.title")}</h2>

      <section className="settings-section" aria-labelledby="settings-about">
        <h3 id="settings-about" className="settings-section-title">
          {t("settings.about")}
        </h3>
        <dl className="settings-meta">
          <div className="settings-meta-row">
            <dt>{t("settings.version")}</dt>
            <dd>
              <code>{info?.version ?? "…"}</code>
            </dd>
          </div>
          <div className="settings-meta-row">
            <dt>{t("settings.crocVersion")}</dt>
            <dd>
              <code>{info?.crocVersion ?? "…"}</code>
            </dd>
          </div>
          <div className="settings-meta-row">
            <dt>{t("settings.relayActive")}</dt>
            <dd className="settings-meta-relay" title={relaySummary}>
              {s.relayAddress?.trim() ? (
                <code>{s.relayAddress.trim()}</code>
              ) : (
                <span className="hint">—</span>
              )}
              {s.relayAddress6?.trim() ? (
                <code className="settings-meta-relay6">{s.relayAddress6.trim()}</code>
              ) : null}
            </dd>
          </div>
          {info?.defaultRelay ? (
            <div className="settings-meta-row">
              <dt>{t("settings.relayDefault")}</dt>
              <dd className="settings-meta-relay">
                <code>{info.defaultRelay}</code>
                {info.defaultRelay6 ? (
                  <code className="settings-meta-relay6">{info.defaultRelay6}</code>
                ) : null}
              </dd>
            </div>
          ) : null}
        </dl>
      </section>

      <section className="settings-section" aria-labelledby="settings-relay">
        <h3 id="settings-relay" className="settings-section-title">
          {t("settings.relaySection")}
        </h3>
        <p className="hint settings-section-hint">{t("settings.relayHint")}</p>
      </section>

      <div className="grid-2">
        <label className="field">
          <span className="field-label">{t("settings.relayV4")}</span>
          <input
            className="input"
            value={s.relayAddress}
            onChange={(e) => update("relayAddress", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.relayV6")}</span>
          <input
            className="input"
            value={s.relayAddress6}
            onChange={(e) => update("relayAddress6", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.relayPassword")}</span>
          <input
            className="input"
            type="password"
            value={s.relayPassword}
            onChange={(e) => update("relayPassword", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.downloadDir")}</span>
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
          <span className="field-label">{t("settings.socks5")}</span>
          <input
            className="input"
            placeholder={t("settings.proxyPlaceholder")}
            value={s.socks5}
            onChange={(e) => update("socks5", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.httpProxy")}</span>
          <input
            className="input"
            placeholder={t("settings.proxyPlaceholder")}
            value={s.httpProxy}
            onChange={(e) => update("httpProxy", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.exclude")}</span>
          <input
            className="input"
            placeholder={t("settings.excludePlaceholder")}
            value={s.exclude}
            onChange={(e) => update("exclude", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.throttle")}</span>
          <input
            className="input"
            placeholder={t("settings.throttlePlaceholder")}
            value={s.throttleUpload}
            onChange={(e) => update("throttleUpload", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.ip")}</span>
          <input
            className="input"
            placeholder={t("settings.ipPlaceholder")}
            value={s.ip}
            onChange={(e) => update("ip", e.target.value)}
          />
        </label>
        <label className="field">
          <span className="field-label">{t("settings.theme")}</span>
          <select
            className="input"
            value={s.theme}
            onChange={(e) => update("theme", e.target.value as Settings["theme"])}
          >
            <option value="system">{t("settings.themeSystem")}</option>
            <option value="light">{t("settings.themeLight")}</option>
            <option value="dark">{t("settings.themeDark")}</option>
          </select>
        </label>
        <label className="field">
          <span className="field-label">{t("settings.language")}</span>
          <select
            className="input"
            value={s.language}
            onChange={(e) => update("language", e.target.value)}
          >
            <option value="system">{t("settings.langSystem")}</option>
            {availableLanguages.map((l) => (
              <option key={l} value={l}>
                {languageNames[l] ?? l}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span className="field-label">{t("settings.curve")}</span>
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
          <span className="field-label">{t("settings.hash")}</span>
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
            ["onlyLocal", t("settings.onlyLocal")],
            ["disableLocal", t("settings.disableLocal")],
            ["noCompress", t("settings.noCompress")],
            ["overwrite", t("settings.overwrite")],
            ["zipFolder", t("settings.zipFolder")],
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
        {saved ? t("settings.saved") : t("settings.save")}
      </button>
    </div>
  );
}

