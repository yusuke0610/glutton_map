import { afterEach, describe, expect, it, vi } from "vitest";
import { createPin, fetchPins } from "./api";

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

describe("createPin", () => {
  it("POST /api/pins に JSON を送り、作成されたピンを返す", async () => {
    const payload = {
      nickname: "如月ファン",
      prefecture: "高知県" as const,
      city: "高知市",
      comment: "唐揚げ最高",
    };
    const created = { ...payload, lat: 33.56, lng: 133.53 };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => created,
    });
    vi.stubGlobal("fetch", fetchMock);

    const got = await createPin(payload);

    expect(got).toEqual(created);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toContain("/api/pins");
    expect(init.method).toBe("POST");
    expect(init.headers["Content-Type"]).toContain("application/json");
    expect(JSON.parse(init.body)).toEqual(payload);
  });

  it("レスポンスが ok でなければ例外を投げる", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: false, status: 400 }),
    );

    await expect(
      createPin({ nickname: "x", prefecture: "高知県", city: "高知市" }),
    ).rejects.toThrow("createPin failed: 400");
  });
});
