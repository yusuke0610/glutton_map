import { defineConfig, devices } from "@playwright/test";

// E2E は backend(:8000) と frontend(:5173) を両方起動して縦割りスライスを通す。
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: "list",
  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: [
    {
      // バックエンド。E2E 用の使い捨て DB に seed が流れる。
      command:
        "cd ../backend && LIBSQL_URL=file:./e2e.db PORT=8000 go run ./cmd/server",
      url: "http://localhost:8000/api/pins",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      // フロント開発サーバ。
      command: "bun run dev",
      url: "http://localhost:5173",
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
  ],
});
