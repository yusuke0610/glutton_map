import { afterEach, describe, expect, it, vi } from "vitest";
import { fetchPins } from "./api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("fetchPins", () => {
  it("成功時はパースした JSON を返す", async () => {
    const data = { pins: [], prefecture_count: 0, total: 0 };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => data,
    });
    vi.stubGlobal("fetch", fetchMock);

    const got = await fetchPins();

    expect(got).toEqual(data);
    // /api/pins を叩いていること。
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining("/api/pins"));
  });

  it("レスポンスが ok でなければ例外を投げる", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: false, status: 500 }),
    );

    await expect(fetchPins()).rejects.toThrow("fetchPins failed: 500");
  });
});
