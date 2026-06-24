import { describe, expect, it } from "vitest";
import {
  buildShareUrl,
  buildTweetIntentUrl,
  defaultShareText,
} from "./share";

describe("buildShareUrl", () => {
  it("backend の /share を指し UTM を付ける（流入を計測できる）", () => {
    const u = buildShareUrl("https://api.example.com", "高知県");
    const parsed = new URL(u);
    expect(parsed.origin + parsed.pathname).toBe("https://api.example.com/share");
    expect(parsed.searchParams.get("pref")).toBe("高知県");
    expect(parsed.searchParams.get("utm_source")).toBe("twitter");
    expect(parsed.searchParams.get("utm_medium")).toBe("social");
    expect(parsed.searchParams.get("utm_campaign")).toBe("fan_share");
  });

  it("末尾スラッシュを二重にしない", () => {
    const u = buildShareUrl("https://api.example.com/", "高知県");
    expect(u).not.toContain("//share");
  });

  it("県が無くても UTM 付きの共有 URL を作る", () => {
    const u = buildShareUrl("https://api.example.com");
    const parsed = new URL(u);
    expect(parsed.searchParams.get("pref")).toBeNull();
    expect(parsed.searchParams.get("utm_source")).toBe("twitter");
  });
});

describe("buildTweetIntentUrl", () => {
  it("intent/tweet に text と url を入れる", () => {
    const u = buildTweetIntentUrl("こんにちは", "https://api.example.com/share?pref=高知県");
    const parsed = new URL(u);
    expect(parsed.origin + parsed.pathname).toBe("https://twitter.com/intent/tweet");
    expect(parsed.searchParams.get("text")).toBe("こんにちは");
    expect(parsed.searchParams.get("url")).toBe("https://api.example.com/share?pref=高知県");
  });
});

describe("defaultShareText", () => {
  it("県名を含む編集可能なデフォルト文面", () => {
    expect(defaultShareText("高知県")).toContain("高知県");
  });
  it("県が無くても壊れない", () => {
    expect(defaultShareText()).not.toBe("");
  });
});
