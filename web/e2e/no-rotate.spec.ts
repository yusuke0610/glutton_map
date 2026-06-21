import { expect, test } from "@playwright/test";

declare global {
  interface Window {
    __map: import("maplibre-gl").Map;
  }
}

// 北を常に上に固定したい（上=ロシア / 下=南極の向きを保つ）。
// ドラッグで地図の向き（bearing）や傾き（pitch）が変わらないことを保証する。
test("ドラッグしても地図が回転・チルトしない（北が上に固定）", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator("#map canvas").first()).toBeVisible({
    timeout: 15_000,
  });

  const box = await page.locator("#map canvas").first().boundingBox();
  if (!box) throw new Error("canvas not found");
  const cx = box.x + box.width / 2;
  const cy = box.y + box.height / 2;

  // 右ドラッグ（MapLibre 既定の回転ジェスチャ）。
  await page.mouse.move(cx, cy);
  await page.mouse.down({ button: "right" });
  await page.mouse.move(cx + 160, cy + 40, { steps: 12 });
  await page.mouse.up({ button: "right" });

  // Ctrl+ドラッグ（既定のチルトジェスチャ）。
  await page.keyboard.down("Control");
  await page.mouse.move(cx, cy);
  await page.mouse.down();
  await page.mouse.move(cx, cy - 140, { steps: 12 });
  await page.mouse.up();
  await page.keyboard.up("Control");

  const bearing = await page.evaluate(() => window.__map.getBearing());
  const pitch = await page.evaluate(() => window.__map.getPitch());
  // -0 になり得るので近接比較する（回転・チルトが起きていない＝実質 0）。
  expect(bearing).toBeCloseTo(0, 5);
  expect(pitch).toBeCloseTo(0, 5);
});
