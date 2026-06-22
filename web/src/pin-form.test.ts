import { describe, it, expect } from "vitest";
import { canSubmitPin } from "./pin-form";

describe("canSubmitPin", () => {
  it("都道府県と市区町村コードが揃い投稿中でなければ投稿できる", () => {
    expect(
      canSubmitPin({ prefecture: "高知県", municipalityCode: "39201", submitting: false }),
    ).toBe(true);
  });

  it("都道府県が未選択なら投稿できない", () => {
    expect(
      canSubmitPin({ prefecture: "", municipalityCode: "39201", submitting: false }),
    ).toBe(false);
  });

  it("市区町村コードが未選択（候補未選択の自由入力）なら投稿できない", () => {
    expect(
      canSubmitPin({ prefecture: "高知県", municipalityCode: "", submitting: false }),
    ).toBe(false);
  });

  it("投稿中は投稿できない（二重送信防止）", () => {
    expect(
      canSubmitPin({ prefecture: "高知県", municipalityCode: "39201", submitting: true }),
    ).toBe(false);
  });
});
