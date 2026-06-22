// ピン投稿時の「手で打ち込む」演出に関する純粋ロジック。
// DOM/MapLibre に依存する副作用は App.tsx 側に置き、ここはテスト可能な判定・定数だけを持つ。

// 演出のタイミング（ミリ秒）。落下完了 = 着地（ピン出現）とし、その後に小休止してからズームする。
export const DROP_TIMING = {
  // ピンが画面上から落下しきるまで。
  dropMs: 420,
  // 着地（ピンが地図に出現する）タイミング。落下完了と同時。
  impactMs: 420,
  // 着地後、ズーム開始までの小休止（刺さったピンを一拍見せる）。
  settleMs: 280,
} as const;

// 投稿後にカメラを寄せる先。市区町村が見える程度（zoom 12）まで flyTo する。
const FLY_TO_ZOOM = 12;
const FLY_TO_DURATION_MS = 1200;

// prefers-reduced-motion が有効なら手の演出はスキップする（ズーム＋popup だけ実行）。
export function shouldAnimateDrop(prefersReducedMotion: boolean): boolean {
  return !prefersReducedMotion;
}

// 投稿地点へズームインする flyTo オプション。マジックナンバーをここに集約する。
export function flyToOptionsFor(
  lng: number,
  lat: number,
): { center: [number, number]; zoom: number; duration: number } {
  return { center: [lng, lat], zoom: FLY_TO_ZOOM, duration: FLY_TO_DURATION_MS };
}
