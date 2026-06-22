// 投稿フォームの送信可否を判定する純粋関数。
// 市区町村はあいまい検索の候補から選択された（municipalityCode が空でない）ときだけ投稿を許す。
// 自由入力のまま候補を選んでいない場合は、サーバ側でも 400 になるため、ここで弾いて即フィードバックする。
export function canSubmitPin(args: {
  prefecture: string;
  municipalityCode: string;
  submitting: boolean;
}): boolean {
  const { prefecture, municipalityCode, submitting } = args;
  return prefecture !== "" && municipalityCode !== "" && !submitting;
}
