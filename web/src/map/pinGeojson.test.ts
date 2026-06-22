import { describe, it, expect } from "vitest";
import { buildPinGeojson } from "./pinGeojson";
import type { Pin } from "../api/api";

describe("buildPinGeojson", () => {
  it("ピンを Point Feature の FeatureCollection に変換する", () => {
    const pins: Pin[] = [
      {
        prefecture: "東京都",
        lat: 35.68,
        lng: 139.76,
        weight: 1,
        nickname: "ぐると",
        city: "千代田区",
        comment: "うまい",
      },
    ];

    const fc = buildPinGeojson(pins);

    expect(fc.type).toBe("FeatureCollection");
    expect(fc.features).toHaveLength(1);
    const f = fc.features[0];
    expect(f.geometry).toEqual({ type: "Point", coordinates: [139.76, 35.68] });
    expect(f.properties).toEqual({
      weight: 1,
      prefecture: "東京都",
      nickname: "ぐると",
      city: "千代田区",
      comment: "うまい",
    });
  });

  it("weight 未指定は 1、欠けた表示項目は空文字で埋める（seed 由来ピン相当）", () => {
    // weight を欠いた seed 由来相当のピン（型上は number だが実データで欠ける場合に備える）。
    const pins = [
      { prefecture: "北海道", lat: 43.06, lng: 141.35 },
    ] as unknown as Pin[];

    const f = buildPinGeojson(pins).features[0];

    expect(f.properties).toEqual({
      weight: 1,
      prefecture: "北海道",
      nickname: "",
      city: "",
      comment: "",
    });
  });

  it("空配列は features 0 件の FeatureCollection を返す", () => {
    const fc = buildPinGeojson([]);
    expect(fc.type).toBe("FeatureCollection");
    expect(fc.features).toHaveLength(0);
  });
});
