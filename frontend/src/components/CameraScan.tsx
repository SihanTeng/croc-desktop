import { useCallback, useEffect, useRef, useState } from "react";
import { App as Backend } from "../api";
import { useT } from "../i18n";

function errMsg(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}

/**
 * Full-screen camera overlay that continuously captures frames and sends them
 * to the Go QR decoder (DecodeCodeFromBase64) until a code is found or the
 * user cancels.
 */
export default function CameraScan({
  onCode,
  onClose,
}: {
  onCode: (code: string) => void;
  onClose: () => void;
}) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const busyRef = useRef(false);
  const stoppedRef = useRef(false);
  const [error, setError] = useState<string | null>(null);
  const [hint, setHint] = useState<string | null>(null);
  const t = useT();

  const stop = useCallback(() => {
    stoppedRef.current = true;
    streamRef.current?.getTracks().forEach((tr) => tr.stop());
    streamRef.current = null;
    if (videoRef.current) videoRef.current.srcObject = null;
  }, []);

  const close = useCallback(() => {
    stop();
    onClose();
  }, [onClose, stop]);

  // Open the camera once; prefer the rear lens on phones.
  useEffect(() => {
    stoppedRef.current = false;
    let cancelled = false;

    (async () => {
      if (!navigator.mediaDevices?.getUserMedia) {
        setError(t("receive.cameraUnsupported"));
        return;
      }
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: false,
          video: {
            facingMode: { ideal: "environment" },
            width: { ideal: 1280 },
            height: { ideal: 720 },
          },
        });
        if (cancelled || stoppedRef.current) {
          stream.getTracks().forEach((tr) => tr.stop());
          return;
        }
        streamRef.current = stream;
        const video = videoRef.current;
        if (video) {
          video.srcObject = stream;
          await video.play();
        }
        setHint(t("receive.cameraHint"));
      } catch (e) {
        if (!cancelled) setError(errMsg(e) || t("receive.cameraDenied"));
      }
    })();

    return () => {
      cancelled = true;
      stop();
    };
  }, [stop, t]);

  // Poll frames while the stream is live.
  useEffect(() => {
    if (error) return;

    const tick = async () => {
      if (stoppedRef.current || busyRef.current) return;
      const video = videoRef.current;
      if (!video || video.readyState < 2 || video.videoWidth < 16) return;

      busyRef.current = true;
      try {
        const canvas = document.createElement("canvas");
        // Downscale large frames — decode is CPU-heavy and QR codes stay readable.
        const maxW = 960;
        const scale = video.videoWidth > maxW ? maxW / video.videoWidth : 1;
        canvas.width = Math.max(1, Math.round(video.videoWidth * scale));
        canvas.height = Math.max(1, Math.round(video.videoHeight * scale));
        const ctx = canvas.getContext("2d");
        if (!ctx) return;
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
        const dataUrl = canvas.toDataURL("image/jpeg", 0.85);
        const decoded = await Backend.DecodeCodeFromBase64(dataUrl);
        if (stoppedRef.current) return;
        stop();
        onCode(decoded);
      } catch {
        // no QR in this frame — keep scanning
      } finally {
        busyRef.current = false;
      }
    };

    const id = window.setInterval(tick, 450);
    return () => window.clearInterval(id);
  }, [error, onCode, stop]);

  // Esc closes the scanner
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [close]);

  return (
    <div
      className="camera-overlay"
      role="dialog"
      aria-modal="true"
      aria-label={t("receive.scanCamera")}
    >
      <div className="camera-stage">
        <video ref={videoRef} className="camera-video" playsInline muted autoPlay />
        <div className="camera-frame" aria-hidden="true" />
        {error ? (
          <p className="camera-banner camera-banner-err">{error}</p>
        ) : (
          hint && <p className="camera-banner">{hint}</p>
        )}
      </div>
      <div className="camera-actions">
        <button type="button" className="btn btn-ghost" onClick={close}>
          {t("common.cancel")}
        </button>
      </div>
    </div>
  );
}
