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
  await page.getByLabel(messages.form.city).fill("高知市");
  await page.getByLabel(messages.form.comment).fill("唐揚げ弁当が最高");
}

test("投稿するとピンがマップに反映される", async ({ page }) => {
  await page.goto("/");
  // 初回のピン取得を待つ。
  await page.waitForResponse(
    (r) => r.url().includes("/api/pins") && r.status() === 200,
  );

  // 後で自分の投稿を特定できるよう一意なニックネームにする。
  const nickname = `E2E_${Date.now()}`;
  await fillForm(page, nickname);

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
