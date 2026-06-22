import type { Pin } from "../api/api";

// ピン配列を地図ソース用の GeoJSON に変換する。読み込み時と投稿後の更新で共用する。
export function buildPinGeojson(pins: Pin[]): GeoJSON.FeatureCollection {
  return {
    type: "FeatureCollection",
    features: pins.map((p) => ({
      type: "Feature",
      geometry: { type: "Point", coordinates: [p.lng, p.lat] },
      properties: {
        weight: p.weight ?? 1,
        // ポップアップ表示用。seed 由来のピンは空文字。
        prefecture: p.prefecture ?? "",
        nickname: p.nickname ?? "",
        city: p.city ?? "",
        comment: p.comment ?? "",
      },
    })),
  };
}
