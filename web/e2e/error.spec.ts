import { expect, test } from "@playwright/test";
import { messages } from "../src/messages";

test("API 失敗時はユーザー向けエラーバナーを表示する", async ({ page }) => {
  // /api/pins を 500 にして取得を失敗させる。
  await page.route("**/api/pins", (route) =>
    route.fulfill({ status: 500, body: "" }),
  );

  await page.goto("/");

  // 握りつぶさず、role=alert のバナーに一元管理した文言を出すこと。
  const banner = page.getByRole("alert");
  await expect(banner).toBeVisible();
  await expect(banner).toContainText(messages.error.fetchPins);
});

test("再試行で復帰し地図を描画する", async ({ page }) => {
  // 最初の1回だけ失敗させ、以降は本物のバックエンドへ通す。
  let failed = false;
  await page.route("**/api/pins", (route) => {
    if (!failed) {
      failed = true;
      return route.fulfill({ status: 500, body: "" });
    }
    return route.continue();
  });

  await page.goto("/");
  await expect(page.getByRole("alert")).toBeVisible();

  // 「再試行」でバナーが消え、地図 canvas が描画される。
  await page.getByRole("button", { name: "再試行" }).click();
  await expect(page.getByRole("alert")).toBeHidden();
  await expect(page.locator("#map canvas").first()).toBeVisible({
    timeout: 15_000,
  });
});
