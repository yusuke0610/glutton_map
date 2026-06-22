import { describe, expect, it } from "vitest";
import { counterText, formatCount } from "./counter";
import { messages } from "./messages";

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
