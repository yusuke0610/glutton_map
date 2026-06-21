import { describe, expect, it } from "vitest";
import { popupHTML } from "./popup";

describe("popupHTML", () => {
  it("ニックネーム・都道府県・市区町村・コメントを含む", () => {
    const html = popupHTML({
      nickname: "如月ファン",
      prefecture: "高知県",
      city: "高知市",
      comment: "唐揚げ最高",
    });
    expect(html).toContain("如月ファン");
    expect(html).toContain("高知県");
    expect(html).toContain("高知市");
    expect(html).toContain("唐揚げ最高");
  });

  it("HTML特殊文字をエスケープして XSS を防ぐ", () => {
    const html = popupHTML({
      nickname: "<script>alert(1)</script>",
      comment: "a & b < c",
    });
    expect(html).not.toContain("<script>");
    expect(html).toContain("&lt;script&gt;");
    expect(html).toContain("a &amp; b &lt; c");
  });

  it("コメント未指定でも壊れない", () => {
    const html = popupHTML({ nickname: "ファン", prefecture: "東京都", city: "渋谷区" });
    expect(html).toContain("ファン");
    expect(typeof html).toBe("string");
  });
});
