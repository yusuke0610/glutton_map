// 匿名投稿者トークン。最初のピンに認証を要求しないため、ブラウザで生成・保持する
// 識別子で「同じファン」を緩く識別し、後からアカウントへ claim する余地を残す。
// localStorage に保存して再訪でも同一トークンを使う。

export const ANON_TOKEN_KEY = "glutton_map.anon_token";

// getOrCreateAnonToken は保存済みトークンを返し、無ければ生成して保存する。
// テスト可能にするため Storage を注入できる（既定はブラウザの localStorage）。
// Storage が使えない環境では保存せずに都度生成したトークンを返す（投稿は止めない）。
export function getOrCreateAnonToken(
  storage: Storage | undefined = globalThis.localStorage,
): string {
  if (!storage) return generateToken();
  const existing = storage.getItem(ANON_TOKEN_KEY);
  if (existing) return existing;
  const token = generateToken();
  storage.setItem(ANON_TOKEN_KEY, token);
  return token;
}

// generateToken は crypto.randomUUID を優先し、無い環境では簡易フォールバックで生成する。
function generateToken(): string {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();
  return `anon-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}
