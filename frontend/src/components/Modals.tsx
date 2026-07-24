import { AcceptPayload, OverwritePayload, formatBytes } from "../api";
import { Modal } from "./Modal";

const MAX_LISTED = 8;

export function AcceptModal({
  payload,
  onRespond,
}: {
  payload: AcceptPayload;
  onRespond: (accept: boolean) => void;
}) {
  const extra = payload.files.length - MAX_LISTED;
  return (
    <Modal
      title={payload.isText ? "Incoming text message" : "Incoming files"}
      footer={
        <>
          <button className="btn btn-ghost" onClick={() => onRespond(false)}>
            Decline
          </button>
          <button className="btn btn-primary" onClick={() => onRespond(true)}>
            Accept
          </button>
        </>
      }
    >
      <p className="modal-sub">
        From <code>{payload.senderId || "unknown"}</code>
        {payload.totalFolders > 0 &&
          ` · ${payload.totalFolders} folder${payload.totalFolders > 1 ? "s" : ""}`}
        {" · "}
        {formatBytes(payload.totalSize)}
      </p>
      {!payload.isText && (
        <ul className="file-list modal-file-list">
          {payload.files.slice(0, MAX_LISTED).map((f, i) => (
            <li key={i}>
              <span className="file-name">
                {f.folder && f.folder !== "." ? `${f.folder}/` : ""}
                {f.name}
              </span>
              <span className="file-size">{formatBytes(f.size)}</span>
            </li>
          ))}
          {extra > 0 && <li className="file-more">…and {extra} more</li>}
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
  const isResume = payload.resumePct < 99;
  return (
    <Modal
      title={isResume ? "Resume transfer?" : "Overwrite file?"}
      footer={
        <>
          <button className="btn btn-ghost" onClick={() => onRespond(false)}>
            Skip
          </button>
          <button className="btn btn-primary" onClick={() => onRespond(true)}>
            {isResume ? "Resume" : "Overwrite"}
          </button>
        </>
      }
    >
      <p className="modal-sub break-all">{payload.path}</p>
      {isResume && (
        <p className="modal-sub">
          {payload.resumePct.toFixed(1)}% is already present locally.
        </p>
      )}
    </Modal>
  );
}
