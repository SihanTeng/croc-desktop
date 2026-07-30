import { useEffect, useMemo, useState } from "react";
import { App as Backend, formatBytes } from "../api";
import { decodeDataUrlText, maxTextPreview, previewKind } from "../filepreview";
import { useT } from "../i18n";

/* Shared file display: type icons, rows, and inline previews.
 * Used by Send (picked files), Receive (done card), History, and the
 * accept modal — one implementation, no per-page reinvention. */

export type FileIconKind =
  "folder" | "image" | "video" | "audio" | "archive" | "doc" | "text" | "file";

const archiveExts = new Set(["zip", "tar", "gz", "tgz", "bz2", "xz", "7z", "rar", "zst"]);
const docExts = new Set(["pdf", "doc", "docx", "odt", "rtf"]);
const textExts = new Set(["txt", "md", "log", "csv"]);

/** map a file name to its icon category (folders pass isDir) */
export function fileIconKind(name: string, isDir = false): FileIconKind {
  if (isDir) return "folder";
  const ext = name.split(".").pop()?.toLowerCase() ?? "";
  const preview = previewKind(name);
  if (preview === "image" || preview === "video" || preview === "audio") return preview;
  if (archiveExts.has(ext)) return "archive";
  if (docExts.has(ext)) return "doc";
  if (textExts.has(ext) || preview === "text") return "text";
  return "file";
}

// simple stroke glyphs in the design-system idiom (currentColor, 1.5 width)
function IconGlyph({ kind }: { kind: FileIconKind }) {
  switch (kind) {
    case "folder":
      return (
        <path d="M2 4.5a1 1 0 0 1 1-1h3l1.5 1.5H13a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1v-7.5Z" />
      );
    case "image":
      return (
        <>
          <rect x="2.5" y="3.5" width="11" height="9" rx="1" />
          <circle cx="6" cy="6.5" r="1" />
          <path d="m4 12 3.5-3.5 2.5 2.5 2-2 2.5 3" />
        </>
      );
    case "video":
      return (
        <>
          <rect x="2.5" y="3.5" width="11" height="9" rx="1" />
          <path d="M6.5 6.2v3.6l3.2-1.8-3.2-1.8Z" />
        </>
      );
    case "audio":
      return (
        <>
          <path d="M6 10.5V4l5-1v6.5" />
          <circle cx="4.5" cy="10.5" r="1.5" />
          <circle cx="9.5" cy="9.5" r="1.5" />
        </>
      );
    case "archive":
      return (
        <>
          <rect x="3" y="5.5" width="10" height="7" rx="1" />
          <path d="M3 5.5h10M8 3v2.5M6.5 3h3" />
        </>
      );
    case "doc":
      return (
        <>
          <path d="M4 2.5h5.5L12 5v7.5a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-9a1 1 0 0 1 1-1Z" />
          <path d="M9 2.5V5h3" />
        </>
      );
    case "text":
      return (
        <>
          <path d="M4 2.5h8a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-9a1 1 0 0 1 1-1Z" />
          <path d="M5.5 6h5M5.5 8.5h5M5.5 11h3" />
        </>
      );
    default:
      return <path d="M4 2.5h8a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-9a1 1 0 0 1 1-1Z" />;
  }
}

/** file type icon; tint follows the surrounding text color */
export function FileTypeIcon({ name, isDir = false }: { name: string; isDir?: boolean }) {
  return (
    <svg
      className="file-icon"
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <IconGlyph kind={fileIconKind(name, isDir)} />
    </svg>
  );
}

/** inline media/text preview loaded from the backend as a data: URL */
export function FilePreview({ path, name }: { path: string; name: string }) {
  const t = useT();
  const kind = previewKind(name);
  const [dataUrl, setDataUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!kind) return;
    let live = true;
    Backend.GetFileDataURL(path)
      .then((u) => live && setDataUrl(u))
      .catch(() => live && setFailed(true));
    return () => {
      live = false;
    };
  }, [path, kind]);

  if (failed) return <p className="hint recv-name">{t("receive.previewUnavailable")}</p>;
  if (kind === "image" && dataUrl) return <img className="recv-media" src={dataUrl} alt={name} />;
  if (kind === "video" && dataUrl) return <video className="recv-media" src={dataUrl} controls />;
  if (kind === "audio" && dataUrl) return <audio className="recv-audio" src={dataUrl} controls />;
  if (kind === "text" && dataUrl) return <TextPreview dataUrl={dataUrl} />;
  return null;
}

function TextPreview({ dataUrl }: { dataUrl: string }) {
  const text = useMemo(() => decodeDataUrlText(dataUrl), [dataUrl]);
  return <pre className="received-text">{text}</pre>;
}

/** can this row offer a preview? (large text files are listed, not previewed) */
function previewable(name: string, path?: string, size?: number): boolean {
  if (!path) return false;
  const kind = previewKind(name);
  if (!kind) return false;
  if (kind === "text" && size != null && size > maxTextPreview) return false;
  return true;
}

/** trash-can glyph for the remove action */
function TrashIcon() {
  return (
    <svg
      width="13"
      height="13"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M2.5 4h11M6.5 4V2.8a.8.8 0 0 1 .8-.8h1.4a.8.8 0 0 1 .8.8V4M3.5 4l.7 9a1 1 0 0 0 1 .9h5.6a1 1 0 0 0 1-.9l.7-9M6.5 7v4M9.5 7v4" />
    </svg>
  );
}

/** One shared file row: type icon, name, size, optional remove button, and an
 * expandable inline preview when a path is available. */
export function FileRow({
  name,
  path,
  size,
  isDir = false,
  onRemove,
  defaultOpen = false,
  expandable = true,
}: {
  name: string;
  path?: string;
  size?: number;
  isDir?: boolean;
  onRemove?: () => void;
  defaultOpen?: boolean;
  expandable?: boolean;
}) {
  const canPreview = expandable && previewable(name, path, size);
  const [open, setOpen] = useState(defaultOpen && canPreview);
  return (
    <div className="file-row-wrap">
      <div className="file-row">
        <FileTypeIcon name={name} isDir={isDir} />
        <span className="file-name break-all" title={path ?? name}>
          {name}
        </span>
        {size != null && <span className="file-size">{formatBytes(size)}</span>}
        {canPreview && (
          <button
            className={`btn btn-ghost btn-sm file-row-toggle${open ? " file-row-toggle-open" : ""}`}
            onClick={() => setOpen((o) => !o)}
            aria-label="Toggle preview"
          >
            ▸
          </button>
        )}
        {onRemove && (
          <button className="btn btn-ghost btn-sm" onClick={onRemove} aria-label="Delete">
            <TrashIcon />
          </button>
        )}
      </div>
      {open && canPreview && path && (
        <div className="file-row-preview">
          <FilePreview path={path} name={name} />
        </div>
      )}
    </div>
  );
}
