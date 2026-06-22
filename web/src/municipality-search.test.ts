import { describe, it, expect } from "vitest";
import { searchMunicipalities } from "./municipality-search";
import type { Municipality } from "./municipalities";

const LIST: Municipality[] = [
  { code: "13120", prefecture: "東京都", name: "練馬区", kana: "ねりまく" },
  { code: "13112", prefecture: "東京都", name: "世田谷区", kana: "せたがやく" },
  { code: "39201", prefecture: "高知県", name: "高知市", kana: "こうちし" },
];

const names = (ms: Municipality[]) => ms.map((m) => m.name);

describe("searchMunicipalities", () => {
  it("漢字の部分一致でヒットする（練馬→練馬区）", () => {
    expect(names(searchMunicipalities(LIST, "東京都", "練馬"))).toEqual(["練馬区"]);
  });

  it("ひらがな（かな）でヒットする", () => {
    expect(names(searchMunicipalities(LIST, "東京都", "ねりま"))).toEqual(["練馬区"]);
  });

  it("カタカナ入力もひらがなに正規化してヒットする", () => {
    expect(names(searchMunicipalities(LIST, "東京都", "ネリマ"))).toEqual(["練馬区"]);
  });

  it("選択中の都道府県で絞り込む", () => {
    // 高知県を選んでいれば東京の練馬は出ない
    expect(searchMunicipalities(LIST, "高知県", "練馬")).toEqual([]);
  });

  it("都道府県未選択なら全国から検索する", () => {
    expect(names(searchMunicipalities(LIST, "", "高知"))).toEqual(["高知市"]);
  });

  it("空クエリは候補なし", () => {
    expect(searchMunicipalities(LIST, "東京都", "  ")).toEqual([]);
  });

  it("limit で件数を制限する", () => {
    expect(searchMunicipalities(LIST, "東京都", "区", 1)).toHaveLength(1);
  });
});
