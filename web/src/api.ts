import type { components } from "./types.gen";

export type PinsResponse = components["schemas"]["PinsResponse"];

const API_BASE = "http://localhost:8000";

export async function fetchPins(): Promise<PinsResponse> {
  const res = await fetch(`${API_BASE}/api/pins`);
  if (!res.ok) {
    throw new Error(`fetchPins failed: ${res.status}`);
  }
  return res.json();
}
