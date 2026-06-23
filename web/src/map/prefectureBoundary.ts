import type { FilterSpecification, LineLayerSpecification } from "maplibre-gl";

// 都道府県境界の GeoJSON ソース／レイヤーの id。App.tsx で addSource / addLayer する。
export const PREFECTURE_SOURCE_ID = "prefecture-boundary";
export const PREFECTURE_LINE_LAYER_ID = "prefecture-boundary-line";

// 境界 GeoJSON の配信 URL。web/public 配下を Vite が配信する（base 既定 "/"）。
export const PREFECTURE_DATA_URL = `${import.meta.env.BASE_URL}prefectures.geojson`;

// どの県にも一致しない filter（初期状態＝非表示）。実在しない値を比較に使う。
export const NO_PREFECTURE_FILTER: FilterSpecification = [
  "==",
  ["get", "prefecture"],
  "__none__",
];

// boundaryFilterFor は選択中の県だけ線を表示する filter を返す。null なら非表示。
export function boundaryFilterFor(prefecture: string | null): FilterSpecification {
  if (!prefecture) return NO_PREFECTURE_FILTER;
  return ["==", ["get", "prefecture"], prefecture];
}

// prefectureBoundaryLayer は選択中の都道府県を赤線で囲う line レイヤー。
// 初期は非表示 filter にしておき、クリックで boundaryFilterFor により該当県へ切り替える。
export function prefectureBoundaryLayer(): LineLayerSpecification {
  return {
    id: PREFECTURE_LINE_LAYER_ID,
    type: "line",
    source: PREFECTURE_SOURCE_ID,
    filter: NO_PREFECTURE_FILTER,
    layout: { "line-join": "round", "line-cap": "round" },
    paint: { "line-color": "#e0202a", "line-width": 3 },
  };
}
