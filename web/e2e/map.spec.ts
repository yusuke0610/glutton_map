import { expect, test } from "@playwright/test";

test("地図ページが /api/pins を取得して地図を描画する", async ({ page }) => {
  // ページ遷移と同時に /api/pins の 200 応答を待つ。
  const pinsResponse = page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.status() === 200,
  );

  await page.goto("/");

  // バックエンドが seed 済みのピンと集計を返していること。
  const body = await (await pinsResponse).json();
  expect(body.total).toBeGreaterThan(0);
  expect(body.prefecture_count).toBeGreaterThan(0);
  expect(Array.isArray(body.pins)).toBe(true);

  // MapLibre は #map 内の canvas に描画する。canvas が表示されれば地図初期化成功。
  await expect(page.locator("#map canvas").first()).toBeVisible({
    timeout: 15_000,
  });
});
