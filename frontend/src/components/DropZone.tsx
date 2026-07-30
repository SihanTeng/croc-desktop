import { ReactNode, useState } from "react";

// DropZone is a visual drop target. The actual dropped paths arrive through
// the backend "files:dropped" event (native Wails file drop); this component
// handles the highlight state and the idle content. The data-file-drop-target
// attribute marks it as a valid target for Wails v3.
export default function DropZone({ children }: { children: ReactNode }) {
  const [active, setActive] = useState(false);
  return (
    <div
      className={`dropzone ${active ? "dropzone-active" : ""}`}
      data-file-drop-target
      onDragEnter={(e) => {
        e.preventDefault();
        setActive(true);
      }}
      onDragOver={(e) => e.preventDefault()}
      onDragLeave={() => setActive(false)}
      onDrop={(e) => e.preventDefault()}
    >
      {children}
    </div>
  );
}
