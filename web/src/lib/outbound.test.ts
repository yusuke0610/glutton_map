import { describe, expect, it } from "vitest";
import { OUTBOUND, outboundUrl } from "./outbound";

describe("outboundUrl", () => {
  it("backend の /out を to 付きで指す（計測経由で送客）", () => {
    const u = outboundUrl("https://api.example.com", OUTBOUND.menu);
    expect(u).toBe("https://api.example.com/out?to=official_menu");
  });

  it("末尾スラッシュを二重にしない", () => {
    expect(outboundUrl("https://api.example.com/", OUTBOUND.menu)).toBe(
      "https://api.example.com/out?to=official_menu",
    );
  });

  it("to をエスケープする", () => {
    const u = outboundUrl("https://api.example.com", "a b");
    expect(u).toContain("to=a%20b");
  });
});
