import type { StyleSpecification } from "maplibre-gl";

// ベース地図は OpenStreetMap の標準ラスタタイル（完成スタイル）。
// 道路・国道番号・鉄道・駅・地名ラベルはタイル側で描画済みのため、layers はラスタ1枚のみ。
export const SOURCE_ID = "osm";

export const mapStyle: StyleSpecification = {
  version: 8,
  sources: {
    [SOURCE_ID]: {
      type: "raster",
      tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"],
      tileSize: 256,
      maxzoom: 19, // OSM のネイティブ最大ズーム。これ以上は overzoom（拡大）になる。
      // 出典表示は OSM タイル利用ポリシーで必須。MapLibre が右下に表示する。
      attribution:
        '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    },
  },
  layers: [{ id: "osm", type: "raster", source: SOURCE_ID }],
};
