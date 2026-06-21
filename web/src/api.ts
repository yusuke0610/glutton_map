import type { components } from "./types.gen";
import { logger } from "./logger";

export type PinsResponse = components["schemas"]["PinsResponse"];
export type Pin = components["schemas"]["Pin"];
export type CreatePinRequest = components["schemas"]["CreatePinRequest"];

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

// createPin はファン投稿を1件 POST し、作成されたピンを返す。
// 座標はサーバが生成するため payload に lat/lng は含めない。
export async function createPin(payload: CreatePinRequest): Promise<Pin> {
  logger.debug("createPin: posting", payload);
  const res = await fetch(`${API_BASE}/api/pins`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    logger.error("createPin failed", res.status);
    throw new Error(`createPin failed: ${res.status}`);
  }
  return res.json();
}
