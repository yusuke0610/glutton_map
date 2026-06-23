import { afterEach, describe, expect, it, vi } from "vitest";
import { createPin, fetchPins, fetchPrefectureAt } from "./api";

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
      municipality_code: "39201",
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
      createPin({
        nickname: "x",
        prefecture: "高知県",
        city: "高知市",
        municipality_code: "39201",
      }),
    ).rejects.toThrow("createPin failed: 400");
  });
});

describe("fetchPrefectureAt", () => {
  it("200 のときは都道府県と件数を返し、lat/lng をクエリに乗せる", async () => {
    const stat = { prefecture: "東京都" as const, count: 42 };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => stat,
    });
    vi.stubGlobal("fetch", fetchMock);

    const got = await fetchPrefectureAt(35.69, 139.69);

    expect(got).toEqual(stat);
    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain("/api/prefectures/at");
    expect(url).toContain("lat=35.69");
    expect(url).toContain("lng=139.69");
  });

  it("404（海上など）のときは null を返す", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: false, status: 404 }),
    );

    await expect(fetchPrefectureAt(30, 145)).resolves.toBeNull();
  });

  it("その他のエラーは例外を投げる", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: false, status: 500 }),
    );

    await expect(fetchPrefectureAt(35, 139)).rejects.toThrow(
      "fetchPrefectureAt failed: 500",
    );
  });
});
