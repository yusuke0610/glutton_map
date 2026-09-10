import { describe, expect, it } from "vitest";
import { counterText, formatCount, prefectureCoverageText } from "./counter";
import { messages } from "./messages";
import { PREFECTURES } from "../geo/prefectures";

const { prefix, suffix } = messages.counter;

describe("formatCount", () => {
  it("3桁区切りの数字文字列にする", () => {
    expect(formatCount(0)).toBe("0");
    expect(formatCount(12)).toBe("12");
    expect(formatCount(1234)).toBe("1,234");
  });
});

describe("counterText", () => {
  it("総数を文言の prefix と suffix で挟む", () => {
    expect(counterText(12)).toBe(`${prefix}12${suffix}`);
  });

  it("0人でも壊れない", () => {
    const text = counterText(0);
    expect(text).toContain("0");
    expect(text).toContain(prefix);
  });

  it("4桁以上は3桁区切りで読みやすくする", () => {
    expect(counterText(1234)).toBe(`${prefix}1,234${suffix}`);
  });
});

describe("prefectureCoverageText", () => {
  it("N / 47 都道府県 の形式にする", () => {
    const text = prefectureCoverageText(32);
    expect(text).toContain("32");
    expect(text).toContain(String(PREFECTURES.length));
  });

  it("0件でも壊れない", () => {
    expect(prefectureCoverageText(0)).toContain("0");
  });

  it("47件（全都道府県制覇）でも壊れない", () => {
    expect(prefectureCoverageText(47)).toContain("47");
  });
});
