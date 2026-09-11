import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: { port: 5174 },
  // vitest 設定。コンポーネントテスト(*.test.tsx)は DOM が要るため jsdom を使う。
  // E2E（e2e/*.spec.ts, Playwright）を拾わないよう対象を src の単体テストに限定する。
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts", "src/**/*.test.tsx"],
    setupFiles: ["./vitest.setup.ts"],
  },
});
