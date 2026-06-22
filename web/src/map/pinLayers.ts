import type {
  ExpressionSpecification,
  HeatmapLayerSpecification,
  SymbolLayerSpecification,
} from "maplibre-gl";

// pins GeoJSON ソースの id。App.tsx で addSource する。
export const PINS_SOURCE_ID = "pins";
// ピン上部に載せるアイコン画像の id（App.tsx で addImage。今はプレースホルダ）。
export const PIN_ICON_IMAGE = "pin-icon";
// ピン symbol レイヤーの id。クリックでポップアップを出すため App.tsx から参照する。
export const PIN_ICON_LAYER_ID = "pins-icon";

// ハイブリッドの切り替えズーム。
// これ未満 = ヒートマップ（分布が見える）、これ以上 = 食いしんboy ピン（個別の場所）。
// 境界の [Z, Z+1] でクロスフェードする。
export const PIN_TRANSITION_ZOOM = 8;

// ズームインで消えるヒートマップ。配色・強度・半径は従来どおり維持し、
// opacity だけ切り替えズームで 0 までフェードさせる（ピンに主役を譲る）。
export function heatmapLayer(): HeatmapLayerSpecification {
  return {
    id: "pins-heat",
    type: "heatmap",
    source: PINS_SOURCE_ID,
    // 切り替えズーム＋1 以上では描画しない。
    maxzoom: PIN_TRANSITION_ZOOM + 1,
    paint: {
      "heatmap-weight": ["coalesce", ["get", "weight"], 1] as unknown as ExpressionSpecification,
      "heatmap-intensity": [
        "interpolate",
        ["linear"],
        ["zoom"],
        4,
        1,
        9,
        3,
      ] as unknown as ExpressionSpecification,
      "heatmap-radius": [
        "interpolate",
        ["linear"],
        ["zoom"],
        4,
        18,
        9,
        40,
      ] as unknown as ExpressionSpecification,
      "heatmap-color": [
        "interpolate",
        ["linear"],
        ["heatmap-density"],
        0,
        "rgba(217,123,58,0)",
        0.2,
        "rgba(240,217,198,0.6)",
        0.5,
        "rgb(217,123,58)",
        1,
        "rgb(166,71,26)",
      ] as unknown as ExpressionSpecification,
      // 切り替えズーム手前まで 0.85、そこから 0 へフェードアウト。
      "heatmap-opacity": [
        "interpolate",
        ["linear"],
        ["zoom"],
        PIN_TRANSITION_ZOOM,
        0.85,
        PIN_TRANSITION_ZOOM + 1,
        0,
      ] as unknown as ExpressionSpecification,
    },
  };
}

// 切り替えズームから現れる、食いしんboy アイコン付きピンの symbol レイヤー。
export function pinIconLayer(): SymbolLayerSpecification {
  return {
    id: PIN_ICON_LAYER_ID,
    type: "symbol",
    source: PINS_SOURCE_ID,
    minzoom: PIN_TRANSITION_ZOOM,
    layout: {
      "icon-image": PIN_ICON_IMAGE,
      "icon-size": [
        "interpolate",
        ["linear"],
        ["zoom"],
        PIN_TRANSITION_ZOOM,
        0.5,
        12,
        1,
      ] as unknown as ExpressionSpecification,
      // ピン先端（画像下端）が座標を指す。
      "icon-anchor": "bottom",
      // 密集しても全ピンを出す（分布はヒートマップ側で担保）。
      "icon-allow-overlap": true,
    },
    paint: {
      // ヒートマップと入れ替わるよう、切り替えズームで 0→1 にフェードイン。
      "icon-opacity": [
        "interpolate",
        ["linear"],
        ["zoom"],
        PIN_TRANSITION_ZOOM,
        0,
        PIN_TRANSITION_ZOOM + 1,
        1,
      ] as unknown as ExpressionSpecification,
    },
  };
}
