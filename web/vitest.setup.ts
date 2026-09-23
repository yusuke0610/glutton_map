import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";

// テストごとに前回の render() の DOM を除去する（RTL は jest のグローバル afterEach に
// 自動フックするため、vitest では明示的に呼ぶ必要がある）。
afterEach(() => {
  cleanup();
});
