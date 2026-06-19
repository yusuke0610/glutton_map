import type { components } from "./types.gen";

export type PinsResponse = components["schemas"]["PinsResponse"];

// API のベースURL。Vite の環境変数 VITE_API_BASE で上書きでき、
// 未設定ならローカル開発の既定値を使う（ビルド時に値が焼き込まれる）。
const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8000";

export async function fetchPins(): Promise<PinsResponse> {
  const res = await fetch(`${API_BASE}/api/pins`);
  if (!res.ok) {
    throw new Error(`fetchPins failed: ${res.status}`);
  }
  return res.json();
}
