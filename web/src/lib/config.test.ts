import { afterEach, describe, expect, it, vi } from "vitest";

describe("API_BASE", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.resetModules();
  });

  it("VITE_API_BASE が設定されていればその値を使う", async () => {
    vi.stubEnv("VITE_API_BASE", "https://example.com");
    const { API_BASE } = await import("./config");
    expect(API_BASE).toBe("https://example.com");
  });

  it("VITE_API_BASE が未設定なら分かりやすいエラーを投げる（サイレントにlocalhostへフォールバックしない）", async () => {
    vi.stubEnv("VITE_API_BASE", "");
    await expect(import("./config")).rejects.toThrow(/VITE_API_BASE/);
  });
});
