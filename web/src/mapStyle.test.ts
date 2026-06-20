import { describe, expect, it } from "vitest";
import { mapStyle, SOURCE_ID } from "./mapStyle";

describe("mapStyle", () => {
  it("OpenStreetMap の標準ラスタタイルを参照する", () => {
    const src = mapStyle.sources[SOURCE_ID];
    expect(src).toBeDefined();
    expect(src.type).toBe("raster");
    const tiles = "tiles" in src ? src.tiles : undefined;
    // OSM 標準タイル。https であること（http は OSM ポリシーで禁止）。
    expect(tiles?.[0]).toContain("https://tile.openstreetmap.org");
  });

  it("© OpenStreetMap contributors の出典を持つ", () => {
    const src = mapStyle.sources[SOURCE_ID];
    const attribution = ("attribution" in src && src.attribution) || "";
    expect(attribution).toContain("OpenStreetMap");
  });

  it("レイヤーはラスタ1枚のみ（symbol/vector レイヤーを持たない）", () => {
    expect(mapStyle.layers).toHaveLength(1);
    expect(mapStyle.layers[0].type).toBe("raster");
    expect(mapStyle.layers.some((l) => l.type === "symbol")).toBe(false);
  });
});
