import { ReactNode, useState } from "react";

// DropZone is a visual drop target. The actual dropped paths arrive through
// the backend "files:dropped" event (native Wails file drop); this component
// handles the highlight state and the idle content. The CSS custom property
// --wails-drop-target: drop marks it as a valid target for Wails.
export default function DropZone({ children }: { children: ReactNode }) {
  const [active, setActive] = useState(false);
  return (
    <div
      className={`dropzone ${active ? "dropzone-active" : ""}`}
      style={{ ["--wails-drop-target" as any]: "drop" }}
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
