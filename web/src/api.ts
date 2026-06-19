import type { components } from "./types.gen";
import { logger } from "./logger";

export type PinsResponse = components["schemas"]["PinsResponse"];

// API のベースURL。Vite の環境変数 VITE_API_BASE で指定する（ビルド時に焼き込まれる）。
// 値は web/.env 等で設定する前提。
const API_BASE = import.meta.env.VITE_API_BASE;

export async function fetchPins(): Promise<PinsResponse> {
  logger.debug("fetchPins: requesting", `${API_BASE}/api/pins`);
  const res = await fetch(`${API_BASE}/api/pins`);
  if (!res.ok) {
    logger.error("fetchPins failed", res.status);
    throw new Error(`fetchPins failed: ${res.status}`);
  }
  return res.json();
}
