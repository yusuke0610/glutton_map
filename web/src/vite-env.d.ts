/// <reference types="vite/client" />

// Vite の環境変数の型定義。VITE_ 接頭辞のものだけがクライアントに公開される。
interface ImportMetaEnv {
  // API のベースURL（例: http://localhost:8001）。未設定可。
  readonly VITE_API_BASE?: string;
  // ログ出力の最小レベル（debug/info/warn/error）。未設定なら本番warn・開発debug。
  readonly VITE_LOG_LEVEL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
