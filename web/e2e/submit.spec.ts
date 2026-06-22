import { expect, test } from "@playwright/test";
import { messages } from "../src/messages";

declare global {
  interface Window {
    __map: import("maplibre-gl").Map;
  }
}

// フォームを開いて各フィールドを埋めるヘルパー。
async function fillForm(
  page: import("@playwright/test").Page,
  nickname: string,
) {
  await page.getByRole("button", { name: messages.form.open }).click();
  await page.getByLabel(messages.form.nickname).fill(nickname);
  await page.getByLabel(messages.form.prefecture).selectOption("高知県");
  // 市区町村は候補から選択しないと投稿できないため、あいまい検索で高知市を選ぶ。
  await page.getByLabel(messages.form.city).fill("高知市");
  await page.getByRole("button", { name: "高知市" }).click();
  await page.getByLabel(messages.form.comment).fill("唐揚げ弁当が最高");
}

// 注: 投稿には IP 単位 3 秒のクールダウンがあるため、実 POST するテストは1本に保つ
//（並列ワーカーで複数が同時 POST すると 429 になる）。この1本で「反映」と「境界内」を両方検証する。
test("市区町村を候補から選んで投稿するとその境界内にピンが反映される", async ({
  page,
}) => {
  await page.goto("/");
  // 初回のピン取得を待つ。
  await page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.status() === 200,
  );

  // 後で自分の投稿を特定できるよう一意なニックネームにする。
  const nickname = `E2E_${Date.now()}`;
  await page.getByRole("button", { name: messages.form.open }).click();
  await page.getByLabel(messages.form.nickname).fill(nickname);
  await page.getByLabel(messages.form.prefecture).selectOption("東京都");
  // あいまい検索: 「練馬」と打って候補から練馬区を選ぶ。
  await page.getByLabel(messages.form.city).fill("練馬");
  await page.getByRole("button", { name: "練馬区" }).click();

  const posted = page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.request().method() === "POST",
  );
  await page.getByRole("button", { name: messages.form.submit }).click();
  expect((await posted).status()).toBe(201);

  // 成功通知が出ること。
  await expect(page.getByText(messages.form.success)).toBeVisible();

  // マップが再取得され、投稿したニックネームを持つピンがソースに乗ること。
  await expect
    .poll(
      () =>
        page.evaluate((nick) => {
          const feats = window.__map.querySourceFeatures("pins");
          return feats.some((f) => f.properties?.nickname === nick);
        }, nickname),
      { timeout: 15_000 },
    )
    .toBe(true);

  // 投稿ピンの座標が練馬区 bbox（同梱データ）内に入ること。
  const [lng, lat] = (await page.evaluate((nick) => {
    const feats = window.__map.querySourceFeatures("pins");
    const f = feats.find((x) => x.properties?.nickname === nick);
    return (f!.geometry as GeoJSON.Point).coordinates;
  }, nickname)) as [number, number];
  expect(lng).toBeGreaterThanOrEqual(139.56);
  expect(lng).toBeLessThanOrEqual(139.683);
  expect(lat).toBeGreaterThanOrEqual(35.715);
  expect(lat).toBeLessThanOrEqual(35.785);
});

test("候補から市区町村を選んでいないと投稿ボタンが押せない", async ({ page }) => {
  await page.goto("/");
  await page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.status() === 200,
  );

  await page.getByRole("button", { name: messages.form.open }).click();
  await page.getByLabel(messages.form.nickname).fill("候補未選択テスト");
  // 北海道を選んで「高知市」と自由入力（候補には出ない）→ コード未選択。
  await page.getByLabel(messages.form.prefecture).selectOption("北海道");
  await page.getByLabel(messages.form.city).fill("高知市");

  // 候補未選択なので投稿ボタンは無効、ヒントが表示される。
  await expect(page.getByRole("button", { name: messages.form.submit })).toBeDisabled();
  await expect(page.getByText(messages.form.cityHint)).toBeVisible();
});

test("投稿失敗時はフォームにエラーを表示する", async ({ page }) => {
  // POST だけ 500 にし、GET（地図表示）は本物へ通す。
  await page.route("**/api/pins", (route) => {
    if (route.request().method() === "POST") {
      return route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ message: "boom" }),
      });
    }
    return route.continue();
  });

  await page.goto("/");
  await fillForm(page, "失敗テスト");
  await page.getByRole("button", { name: messages.form.submit }).click();

  // 握りつぶさず、一元管理した投稿失敗の文言を出すこと。
  const alert = page.getByRole("alert");
  await expect(alert).toBeVisible();
  await expect(alert).toContainText(messages.error.createPin);
});
