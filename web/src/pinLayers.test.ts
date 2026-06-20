import { describe, expect, it } from "vitest";
import {
  heatmapLayer,
  PIN_ICON_IMAGE,
  PIN_TRANSITION_ZOOM,
  PINS_SOURCE_ID,
  pinIconLayer,
} from "./pinLayers";

describe("heatmapLayer", () => {
  it("pins ソースのヒートマップで、ズームインすると消える", () => {
    const l = heatmapLayer();
    expect(l.type).toBe("heatmap");
    expect(l.source).toBe(PINS_SOURCE_ID);
    // 切り替えズームより上では描かない（ピンに譲る）。
    expect(l.maxzoom ?? Infinity).toBeLessThanOrEqual(PIN_TRANSITION_ZOOM + 1);
    // opacity はズームで 0 までフェードアウトする。
    const op = l.paint?.["heatmap-opacity"];
    expect(Array.isArray(op)).toBe(true);
    expect(op).toContain(0);
  });
});

describe("pinIconLayer", () => {
  it("pins ソースの symbol で、切り替えズームから出る", () => {
    const l = pinIconLayer();
    expect(l.type).toBe("symbol");
    expect(l.source).toBe(PINS_SOURCE_ID);
    expect(l.minzoom).toBe(PIN_TRANSITION_ZOOM);
    expect(l.layout?.["icon-image"]).toBe(PIN_ICON_IMAGE);
    // ピン先端が座標を指すよう下端アンカー。
    expect(l.layout?.["icon-anchor"]).toBe("bottom");
  });

  it("ヒートマップとクロスフェードする（icon-opacity がフェードイン）", () => {
    const op = pinIconLayer().paint?.["icon-opacity"];
    expect(Array.isArray(op)).toBe(true);
    expect(op).toContain(0); // 透明(0)から始まりフェードイン
  });
});
