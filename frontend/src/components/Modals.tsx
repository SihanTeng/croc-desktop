import { AcceptPayload, OverwritePayload, formatBytes } from "../api";
import { Modal } from "./Modal";
import { FileTypeIcon } from "./FileRow";
import { useT } from "../i18n";

const MAX_LISTED = 8;

export function AcceptModal({
  payload,
  onRespond,
}: {
  payload: AcceptPayload;
  onRespond: (accept: boolean) => void;
}) {
  const t = useT();
  const extra = payload.files.length - MAX_LISTED;
  return (
    <Modal
      title={payload.isText ? t("modal.incomingText") : t("modal.incomingFiles")}
      footer={
        <>
          <button className="btn btn-ghost" onClick={() => onRespond(false)}>
            {t("modal.decline")}
          </button>
          <button className="btn btn-primary" onClick={() => onRespond(true)}>
            {t("modal.accept")}
          </button>
        </>
      }
    >
      <p className="modal-sub">
        {t("modal.from")} <code>{payload.senderId || t("modal.unknown")}</code>
        {payload.totalFolders > 0 &&
          ` · ${payload.totalFolders === 1 ? t("modal.folder") : t("modal.folders", { n: payload.totalFolders })}`}
        {" · "}
        {formatBytes(payload.totalSize)}
      </p>
      {!payload.isText && (
        <ul className="file-list modal-file-list">
          {payload.files.slice(0, MAX_LISTED).map((f, i) => (
            <li key={i}>
              <FileTypeIcon name={f.name} />
              <span className="file-name">
                {f.folder && f.folder !== "." ? `${f.folder}/` : ""}
                {f.name}
              </span>
              <span className="file-size">{formatBytes(f.size)}</span>
            </li>
          ))}
          {extra > 0 && <li className="file-more">{t("history.andMore", { n: extra })}</li>}
        </ul>
      )}
    </Modal>
  );
}

export function OverwriteModal({
  payload,
  onRespond,
}: {
  payload: OverwritePayload;
  onRespond: (overwrite: boolean) => void;
}) {
  const t = useT();
  const isResume = payload.resumePct < 99;
  return (
    <Modal
      title={isResume ? t("modal.resumeTitle") : t("modal.overwriteTitle")}
      footer={
        <>
          <button className="btn btn-ghost" onClick={() => onRespond(false)}>
            {t("modal.skip")}
          </button>
          <button className="btn btn-primary" onClick={() => onRespond(true)}>
            {isResume ? t("modal.resume") : t("modal.overwrite")}
          </button>
        </>
      }
    >
      <p className="modal-sub break-all">{payload.path}</p>
      {isResume && (
        <p className="modal-sub">
          {t("modal.percentPresent", { pct: payload.resumePct.toFixed(1) })}
        </p>
      )}
    </Modal>
  );
}
