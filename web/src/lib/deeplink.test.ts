import { describe, expect, it } from "vitest";
import { parseDeepLink } from "./deeplink";

describe("parseDeepLink", () => {
  it("lat/lng/zoom を数値で取り出す", () => {
    const d = parseDeepLink("?lat=33.56&lng=133.53&zoom=10");
    expect(d.lat).toBeCloseTo(33.56);
    expect(d.lng).toBeCloseTo(133.53);
    expect(d.zoom).toBe(10);
  });

  it("post=1 で投稿モード起動フラグが立つ", () => {
    expect(parseDeepLink("?post=1").openForm).toBe(true);
    expect(parseDeepLink("?post=true").openForm).toBe(true);
  });

  it("post が無ければ投稿モードは起動しない", () => {
    expect(parseDeepLink("?lat=33&lng=133").openForm).toBe(false);
    expect(parseDeepLink("").openForm).toBe(false);
  });

  it("utm_* を流入元として取り出す", () => {
    const d = parseDeepLink(
      "?utm_source=twitter&utm_medium=social&utm_campaign=fan_share",
    );
    expect(d.utm).toEqual({
      utm_source: "twitter",
      utm_medium: "social",
      utm_campaign: "fan_share",
    });
  });

  it("utm が無ければ空オブジェクト", () => {
    expect(parseDeepLink("?lat=33").utm).toEqual({});
  });

  it("不正な数値は無視する（NaN を入れない）", () => {
    const d = parseDeepLink("?lat=abc&lng=&zoom=xyz");
    expect(d.lat).toBeUndefined();
    expect(d.lng).toBeUndefined();
    expect(d.zoom).toBeUndefined();
  });

  it("pref を都道府県として取り出す（共有リンクの着地用）", () => {
    expect(parseDeepLink("?pref=高知県").prefecture).toBe("高知県");
    expect(parseDeepLink("").prefecture).toBeUndefined();
  });
});
