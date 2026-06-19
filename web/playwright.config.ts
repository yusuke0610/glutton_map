import { defineConfig, devices } from "@playwright/test";

// E2E は backend(:8001) と frontend(:5174) を両方起動して縦割りスライスを通す。
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: "list",
  use: {
    baseURL: "http://localhost:5174",
    trace: "on-first-retry",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: [
    {
      // バックエンド。E2E 用の使い捨て DB に seed が流れる。
      command:
        "cd ../backend && LIBSQL_URL=file:./e2e.db PORT=8001 go run ./cmd/server",
      url: "http://localhost:8001/api/pins",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      // フロント開発サーバ。個人の web/.env に左右されないよう、E2E では
      // API のベースURL を E2E バックエンドに固定する（シェルの VITE_* は .env より優先）。
      command: "VITE_API_BASE=http://localhost:8001 bun run dev",
      url: "http://localhost:5174",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
  ],
});
