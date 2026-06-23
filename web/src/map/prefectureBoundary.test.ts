import { describe, expect, it } from "vitest";
import {
  boundaryFilterFor,
  NO_PREFECTURE_FILTER,
  PREFECTURE_SOURCE_ID,
  prefectureBoundaryLayer,
} from "./prefectureBoundary";

describe("prefectureBoundaryLayer", () => {
  it("境界ソースの赤い line レイヤーで、初期は非表示 filter", () => {
    const l = prefectureBoundaryLayer();
    expect(l.type).toBe("line");
    expect(l.source).toBe(PREFECTURE_SOURCE_ID);
    // 赤系の線色。
    expect(l.paint?.["line-color"]).toBe("#e0202a");
    // 初期はどの県にも一致しない filter（選択するまで出さない）。
    expect(l.filter).toEqual(NO_PREFECTURE_FILTER);
  });
});

describe("boundaryFilterFor", () => {
  it("県を渡すとその prefecture に一致する filter を返す", () => {
    expect(boundaryFilterFor("東京都")).toEqual([
      "==",
      ["get", "prefecture"],
      "東京都",
    ]);
  });

  it("null なら非表示 filter を返す", () => {
    expect(boundaryFilterFor(null)).toEqual(NO_PREFECTURE_FILTER);
  });
});
