import { describe, it, expect } from "vitest";
import { shouldAnimateDrop, flyToOptionsFor, DROP_TIMING } from "./pin-drop";

describe("shouldAnimateDrop", () => {
  it("視差効果を減らす設定でなければ演出する", () => {
    expect(shouldAnimateDrop(false)).toBe(true);
  });

  it("prefers-reduced-motion のときは演出しない", () => {
    expect(shouldAnimateDrop(true)).toBe(false);
  });
});

describe("flyToOptionsFor", () => {
  it("投稿地点を中心に市区町村が見える寄り（zoom 12）でアニメーション付きで返す", () => {
    const opts = flyToOptionsFor(133.53, 33.56);
    expect(opts.center).toEqual([133.53, 33.56]);
    expect(opts.zoom).toBe(12);
    expect(opts.duration).toBeGreaterThan(0);
  });
});

describe("DROP_TIMING", () => {
  it("着地は落下完了時点で、その後に小休止してからズームする", () => {
    // 落下 → 着地（ピン出現）→ 小休止 → ズーム、の順序を保証する。
    expect(DROP_TIMING.impactMs).toBe(DROP_TIMING.dropMs);
    expect(DROP_TIMING.settleMs).toBeGreaterThan(0);
  });
});
