import { useCallback, useEffect, useState } from "react";
import {
  AcceptPayload,
  DonePayload,
  OverwritePayload,
  ProgressPayload,
  TransferStateName,
  onEvent,
} from "./api";

export type Direction = "send" | "receive";

export type Phase = TransferStateName | "idle" | "done" | "error";

export interface TransferModel {
  direction: Direction | null;
  phase: Phase;
  code: string;
  progress: ProgressPayload | null;
  accept: AcceptPayload | null;
  overwrite: OverwritePayload | null;
  result: DonePayload | null;
  error: string | null;
  begin: (direction: Direction, code: string) => void;
  clearAccept: () => void;
  clearOverwrite: () => void;
  reset: () => void;
}

interface TransferData {
  direction: Direction | null;
  phase: Phase;
  code: string;
  progress: ProgressPayload | null;
  accept: AcceptPayload | null;
  overwrite: OverwritePayload | null;
  result: DonePayload | null;
  error: string | null;
}

const initial: TransferData = {
  direction: null,
  phase: "idle",
  code: "",
  progress: null,
  accept: null,
  overwrite: null,
  result: null,
  error: null,
};

// useTransfer subscribes to the backend transfer events once, at the app
// root. Only one transfer can run at a time, so this is a single global
// model shared by the Send and Receive views.
export function useTransfer(): TransferModel {
  const [t, setT] = useState<TransferData>(initial);

  useEffect(() => {
    const offs = [
      onEvent<TransferStateName>("transfer:state", (state) =>
        setT((p) => ({ ...p, phase: state }))
      ),
      onEvent<ProgressPayload>("transfer:progress", (progress) =>
        setT((p) => ({ ...p, progress, phase: "transferring" }))
      ),
      onEvent<AcceptPayload>("transfer:accept", (accept) =>
        setT((p) => ({ ...p, accept }))
      ),
      onEvent<OverwritePayload>("transfer:overwrite", (overwrite) =>
        setT((p) => ({ ...p, overwrite }))
      ),
      onEvent<DonePayload>("transfer:done", (result) =>
        setT((p) => ({ ...p, result, phase: "done", progress: null }))
      ),
      onEvent<string>("transfer:error", (error) =>
        setT((p) => ({ ...p, error, phase: "error" }))
      ),
    ];
    return () => offs.forEach((off) => off());
  }, []);

  const begin = useCallback((direction: Direction, code: string) => {
    setT({ ...initial, direction, code, phase: "connecting" });
  }, []);
  const clearAccept = useCallback(
    () => setT((p) => ({ ...p, accept: null })),
    []
  );
  const clearOverwrite = useCallback(
    () => setT((p) => ({ ...p, overwrite: null })),
    []
  );
  const reset = useCallback(() => setT(initial), []);

  return { ...t, begin, clearAccept, clearOverwrite, reset };
}
