/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: { port: 5173 },
  // vitest 設定。ロガーのロジックは DOM 不要なので node 環境で十分。
  test: {
    environment: "node",
  },
});
