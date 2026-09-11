// API のベースURL。Vite の環境変数 VITE_API_BASE で指定する（ビルド時に焼き込まれる）。
// 値は web/.env.example を参考に web/.env.local 等へ設定する。
// 未設定のまま localhost 等へサイレントにフォールバックすると、設定漏れに気付かないまま
// `undefined/api/pins` を叩いて原因不明の 404 になる（#53）。設定漏れは早期に検出するため
// 未設定なら分かりやすいエラーで落とす。
export const API_BASE = requireEnv("VITE_API_BASE");

function requireEnv(key: "VITE_API_BASE"): string {
  const value = import.meta.env[key];
  if (!value) {
    throw new Error(
      `${key} が未設定です。web/.env.example を参考に web/.env.local 等へ設定してください。`,
    );
  }
  return value;
}
