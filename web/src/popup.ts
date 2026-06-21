// ピンのポップアップ表示。ユーザー入力（ニックネーム/コメント等）をそのまま
// HTML へ埋め込むため、必ずエスケープして XSS を防ぐ。

export type PinPopupProps = {
  nickname?: string;
  prefecture?: string;
  city?: string;
  comment?: string;
};

// escapeHTML は HTML 特殊文字を実体参照へ置換する。
export function escapeHTML(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

// popupHTML はピン情報からポップアップの HTML 文字列を組み立てる（全値エスケープ済み）。
export function popupHTML(props: PinPopupProps): string {
  const nickname = escapeHTML(props.nickname ?? "ファン");
  const location = [props.prefecture, props.city]
    .filter((v): v is string => !!v)
    .map(escapeHTML)
    .join(" ");

  const parts = [`<div class="pin-popup">`, `<strong>${nickname}</strong>`];
  if (location) {
    parts.push(`<div class="pin-popup__loc">${location}</div>`);
  }
  if (props.comment) {
    parts.push(`<p class="pin-popup__comment">${escapeHTML(props.comment)}</p>`);
  }
  parts.push(`</div>`);
  return parts.join("");
}
