import { describe, expect, it } from "vitest";
import { ANON_TOKEN_KEY, getOrCreateAnonToken } from "./anonToken";

// node 環境には localStorage が無いため、最小のメモリ Storage を注入してテストする。
function memoryStorage(initial: Record<string, string> = {}): Storage {
  const map = new Map<string, string>(Object.entries(initial));
  return {
    get length() {
      return map.size;
    },
    clear: () => map.clear(),
    getItem: (k: string) => (map.has(k) ? (map.get(k) as string) : null),
    key: (i: number) => Array.from(map.keys())[i] ?? null,
    removeItem: (k: string) => map.delete(k),
    setItem: (k: string, v: string) => {
      map.set(k, v);
    },
  } as Storage;
}

describe("getOrCreateAnonToken", () => {
  it("空の Storage には新規トークンを生成して保存する", () => {
    const storage = memoryStorage();
    const token = getOrCreateAnonToken(storage);
    expect(token).not.toBe("");
    expect(storage.getItem(ANON_TOKEN_KEY)).toBe(token);
  });

  it("既存トークンがあれば再利用する（同じファンを識別し続ける）", () => {
    const storage = memoryStorage({ [ANON_TOKEN_KEY]: "existing-token" });
    expect(getOrCreateAnonToken(storage)).toBe("existing-token");
  });

  it("同じ Storage への連続呼び出しで同一トークンを返す", () => {
    const storage = memoryStorage();
    const first = getOrCreateAnonToken(storage);
    const second = getOrCreateAnonToken(storage);
    expect(second).toBe(first);
  });

  it("Storage が使えない環境でも空でないトークンを返す（保存はしない）", () => {
    const token = getOrCreateAnonToken(undefined);
    expect(token).not.toBe("");
  });

  it("getItem/setItem が例外を投げても投稿を止めない（トークンを返す）", () => {
    // Safari プライベートモードや quota 超過を模した、必ず throw する Storage。
    const throwing = {
      getItem: () => {
        throw new Error("SecurityError");
      },
      setItem: () => {
        throw new Error("QuotaExceededError");
      },
      removeItem: () => {},
      clear: () => {},
      key: () => null,
      length: 0,
    } as unknown as Storage;
    const token = getOrCreateAnonToken(throwing);
    expect(token).not.toBe("");
  });
});
