import type { components } from "./types.gen";
import { logger } from "../lib/logger";

export type PinsResponse = components["schemas"]["PinsResponse"];
export type Pin = components["schemas"]["Pin"];
export type CreatePinRequest = components["schemas"]["CreatePinRequest"];
export type PrefectureStat = components["schemas"]["PrefectureStat"];

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

// fetchPrefectureAt はクリック地点(lat/lng)の都道府県とそのピン合計件数を取得する。
// 地点がどの都道府県にも属さない(海上など)場合は 404 が返るので null を返し、呼び出し側は吹き出しを出さない。
export async function fetchPrefectureAt(
  lat: number,
  lng: number,
): Promise<PrefectureStat | null> {
  const res = await fetch(
    `${API_BASE}/api/prefectures/at?lat=${lat}&lng=${lng}`,
  );
  if (res.status === 404) return null;
  if (!res.ok) {
    logger.error("fetchPrefectureAt failed", res.status);
    throw new Error(`fetchPrefectureAt failed: ${res.status}`);
  }
  return res.json();
}
