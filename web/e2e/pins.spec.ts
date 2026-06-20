import { expect, test } from "@playwright/test";

declare global {
  interface Window {
    __map: import("maplibre-gl").Map;
  }
}

// ハイブリッド表示: ズームアウト=ヒートマップ、ズームイン=食いしんboy ピン。
test("ヒートマップとピンの両レイヤーが乗り、ズームインでピンが描画される", async ({
  page,
}) => {
  const pins = page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.status() === 200,
  );
  await page.goto("/");
  await pins;
  await expect(page.locator("#map canvas").first()).toBeVisible({
    timeout: 15_000,
  });

  // 両レイヤーとピンアイコンが登録されていること。
  await expect
    .poll(() =>
      page.evaluate(
        () =>
          !!window.__map.getLayer("pins-heat") &&
          !!window.__map.getLayer("pins-icon") &&
          window.__map.hasImage("pin-icon"),
      ),
    )
    .toBe(true);

  // ズームインするとピン（symbol）が実際に描画される。
  // 取りこぼしを避けるため、実在するピンの座標へ寄せてから数える。
  const pinCount = await page.evaluate(async () => {
    const map = window.__map;
    const feats = map.querySourceFeatures("pins");
    if (feats.length === 0) return -1;
    const g = feats[0].geometry as { coordinates: [number, number] };
    map.jumpTo({ zoom: 13, center: g.coordinates });
    await new Promise<void>((res) => map.once("idle", () => res()));
    return map.queryRenderedFeatures({ layers: ["pins-icon"] }).length;
  });
  expect(pinCount).toBeGreaterThan(0);
});
