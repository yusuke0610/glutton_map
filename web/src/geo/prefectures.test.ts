import { describe, expect, it } from "vitest";
import { PREFECTURES } from "./prefectures";

describe("PREFECTURES", () => {
  it("47都道府県を重複なく持つ", () => {
    expect(PREFECTURES).toHaveLength(47);
    expect(new Set(PREFECTURES).size).toBe(47);
  });
});
