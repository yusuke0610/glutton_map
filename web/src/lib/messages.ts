// ユーザー向け表示文言の一元管理（単一の真実源）。
// 重複・表記ゆれを避けるため、画面に出す文言はここに集約する。
// 一方、開発者向けの throw / log 文言はフロー追従でここに置かず、
// 発生箇所（例: api.ts）に直接書く。将来 i18n が必要になったら、
// この構造を react-i18next 等のキー定義へ発展させられる。
export const messages = {
  // 画面上部中央に出すアプリのタイトル。
  title: "914マップ",
  error: {
    fetchPins: "ピンの取得に失敗しました",
    createPin: "ピンの投稿に失敗しました。時間をおいて再度お試しください。",
  },
  counter: {
    // 左上のヒーロー表示。総数を prefix と suffix で挟む（例: 世界にくいしんぼ12人！）。
    prefix: "全世界にくいしんぼが",
    suffix: "人！",
  },
  form: {
    title: "あなたもくいしんぼになろう",
    open: "ピンを立てる",
    close: "閉じる",
    nickname: "ニックネーム",
    prefecture: "都道府県",
    city: "市区町村",
    cityHint: "候補から市区町村を選択してください",
    comment: "コメント（任意）",
    submit: "投稿する",
    submitting: "投稿中…",
    success: "投稿しました！マップに反映されます。",
  },
  share: {
    // 投稿直後の X 共有導線。
    heading: "シェアして仲間を増やそう",
    button: "Xでシェアする",
    close: "閉じる",
    textLabel: "投稿文（編集できます）",
  },
  official: {
    cta: "くいしんぼ如月公式サイト🍱",
  },
} as const;
