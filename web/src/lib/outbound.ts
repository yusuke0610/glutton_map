// 公式 URL への計測付き送客リンク。外部リンクは backend の /out 経由にしてクリックを記録し、
// 「マップが公式に N 件送客した」という実績を残す。リダイレクト先は backend 側のホワイトリストで
// 解決するため、フロントは to キーだけを渡す（任意 URL は渡さない＝オープンリダイレクト防止）。

// OUTBOUND は backend のホワイトリストキーと対応する。
export const OUTBOUND = {
  menu: "official_menu",
} as const;

// outboundUrl は backend の /out?to=<key> を組み立てる。
export function outboundUrl(apiBase: string, to: string): string {
  return `${apiBase.replace(/\/+$/, "")}/out?to=${encodeURIComponent(to)}`;
}
