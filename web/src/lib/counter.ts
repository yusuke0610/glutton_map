import { messages } from "./messages";
import { PREFECTURES } from "../geo/prefectures";

// formatCount は総数を3桁区切りの文字列にする（例: 1234 → "1,234"）。
// 数字部分を別要素で色付け表示するため、整形だけを切り出してある。
export function formatCount(total: number): string {
  return total.toLocaleString("en-US");
}

// counterText はユニークファン数からヒーロー指標の平文文言を組み立てる（aria-label 用）。
export function counterText(uniqueFans: number): string {
  return `${messages.counter.prefix}${formatCount(uniqueFans)}${messages.counter.suffix}`;
}

// prefectureCoverageText は「N / 47 都道府県に広がりました」の平文文言を組み立てる。
export function prefectureCoverageText(prefectureCount: number): string {
  return `${prefectureCount} / ${PREFECTURES.length} ${messages.counter.prefectureCoverageSuffix}`;
}
