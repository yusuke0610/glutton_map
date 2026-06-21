import { messages } from "./messages";

// formatCount は総数を3桁区切りの文字列にする（例: 1234 → "1,234"）。
// 数字部分を別要素で色付け表示するため、整形だけを切り出してある。
export function formatCount(total: number): string {
  return total.toLocaleString("en-US");
}

// counterText はピン総数からヒーロー指標の平文文言を組み立てる（aria-label 用）。
export function counterText(total: number): string {
  return `${messages.counter.prefix}${formatCount(total)}${messages.counter.suffix}`;
}
