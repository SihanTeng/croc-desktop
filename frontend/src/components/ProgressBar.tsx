import { ProgressPayload, formatBytes } from "../api";
import { useT } from "../i18n";

export default function ProgressBar({ progress }: { progress: ProgressPayload }) {
  const t = useT();
  const pct =
    progress.bytesTotal > 0 ? Math.min(100, (progress.bytesDone / progress.bytesTotal) * 100) : 0;
  return (
    <div className="progress-wrap">
      <div className="progress-meta">
        <span className="progress-filename">{progress.filename}</span>
        <span className="progress-bytes">
          {formatBytes(progress.bytesDone)} / {formatBytes(progress.bytesTotal)}
        </span>
      </div>
      <div className="progress-track">
        <div className="progress-fill" style={{ width: `${pct}%` }} />
      </div>
      <div className="progress-meta progress-meta-sub">
        <span>{pct.toFixed(1)}%</span>
        {progress.filesTotal > 1 && (
          <span>
            {t("progress.fileOfTotal", {
              done: Math.min(progress.filesDone + 1, progress.filesTotal),
              total: progress.filesTotal,
            })}
          </span>
        )}
      </div>
    </div>
  );
}
