import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: { port: 5173 },
  // vitest 設定。ロガーのロジックは DOM 不要なので node 環境で十分。
  // E2E（e2e/*.spec.ts, Playwright）を拾わないよう対象を src の単体テストに限定する。
  test: {
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
});
