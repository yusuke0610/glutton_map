import { describe, expect, it } from "vitest";
import { shouldLog, thresholdFor } from "./logger";

describe("thresholdFor", () => {
  it("開発(PROD=false)なら debug まで出す", () => {
    expect(thresholdFor({ PROD: false })).toBe("debug");
  });

  it("本番(PROD=true)なら warn 以上に絞る", () => {
    expect(thresholdFor({ PROD: true })).toBe("warn");
  });

  it("VITE_LOG_LEVEL が妥当なら本番でも優先する", () => {
    expect(thresholdFor({ PROD: true, VITE_LOG_LEVEL: "debug" })).toBe("debug");
  });

  it("VITE_LOG_LEVEL が未知の値なら無視して既定にフォールバック", () => {
    expect(thresholdFor({ PROD: false, VITE_LOG_LEVEL: "verbose" })).toBe(
      "debug",
    );
  });
});

describe("shouldLog", () => {
  it("閾値と同じレベルは出す", () => {
    expect(shouldLog("warn", "warn")).toBe(true);
  });

  it("閾値より深刻なら出す", () => {
    expect(shouldLog("error", "warn")).toBe(true);
  });

  it("閾値より軽いレベルは出さない", () => {
    expect(shouldLog("info", "warn")).toBe(false);
    expect(shouldLog("debug", "info")).toBe(false);
  });
});
