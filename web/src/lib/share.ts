// X 共有（Web Intent）用の URL・文面ビルダ。純粋関数にしてテスト可能にする。
//
// 共有リンクは backend の /share を指す。X クローラは JS を実行しないため /share が
// OGP/Twitter Card 入り HTML を SSR で返し、人間にはフロントへリダイレクトする。
// URL には UTM を付け、踏んだ人の投稿を inflow_source として計測できるようにする。

// 共有リンクに付ける UTM。X からの流入であることを示す。
const SHARE_UTM = {
  utm_source: "twitter",
  utm_medium: "social",
  utm_campaign: "fan_share",
} as const;

// buildShareUrl は backend の /share を指す UTM 付き共有 URL を作る。
// apiBase は backend の公開 URL（本番は https。VITE_API_BASE）。
export function buildShareUrl(apiBase: string, prefecture?: string): string {
  const params = new URLSearchParams();
  if (prefecture) params.set("pref", prefecture);
  for (const [k, v] of Object.entries(SHARE_UTM)) params.set(k, v);
  return `${apiBase.replace(/\/+$/, "")}/share?${params.toString()}`;
}

// defaultShareText は編集可能なツイート文面の初期値。
export function defaultShareText(prefecture?: string): string {
  return prefecture
    ? `${prefecture}のくいしんぼ如月ファンに登録した🍱 みんなも刺してみて`
    : `くいしんぼ如月ファンマップにピンを刺した🍱 みんなも刺してみて`;
}

// buildTweetIntentUrl は X の Web Intent URL を作る（投稿文を事前入力）。
export function buildTweetIntentUrl(text: string, shareUrl: string): string {
  const params = new URLSearchParams({ text, url: shareUrl });
  return `https://twitter.com/intent/tweet?${params.toString()}`;
}
