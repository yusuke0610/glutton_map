import { expect, test } from "@playwright/test";

declare global {
  interface Window {
    __map: import("maplibre-gl").Map;
  }
}

// ズームアウトしすぎて地図が極小になるのを防ぐ。Google マップ程度で下限を止める。
test("一定以上はズームアウトできない（minZoom で下限を固定）", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator("#map canvas").first()).toBeVisible({
    timeout: 15_000,
  });

  // 下限より小さいズームを要求しても、minZoom までしか引けないこと。
  const zoom = await page.evaluate(() => {
    window.__map.jumpTo({ zoom: 0 });
    return window.__map.getZoom();
  });
  const minZoom = await page.evaluate(() => window.__map.getMinZoom());

  expect(minZoom).toBeGreaterThanOrEqual(3);
  expect(zoom).toBeGreaterThanOrEqual(3);
});
