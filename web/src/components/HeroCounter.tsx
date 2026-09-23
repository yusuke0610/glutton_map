import { formatCount } from "../lib/counter";
import { messages } from "../lib/messages";
import { PREFECTURES } from "../geo/prefectures";

// ヒーロー指標（左上）。数字部分だけ赤で強調する。
// 1行目: ユニークファン数（人数。連投を畳んだ値。例: 全世界にくいしんぼが◯◯人！）
// 2行目: 都道府県カバレッジ（例: 32 / 47 都道府県に広がりました）
// aria-label は付けず、可視テキストをそのまま読み上げに使う（フォームの
// 「都道府県」ラベルと文言が衝突し getByLabel が複数要素にマッチするのを避けるため）。
export function HeroCounter({
  uniqueFans,
  prefectureCount,
}: {
  uniqueFans: number;
  prefectureCount: number;
}) {
  return (
    <div
      style={{
        position: "absolute",
        top: 12,
        left: 12,
        zIndex: 2,
        padding: "8px 14px",
        background: "rgba(255,255,255,0.95)",
        color: "#d97b3a",
        borderRadius: 10,
        fontWeight: "bold",
        fontSize: 16,
        boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
        pointerEvents: "none",
      }}
    >
      <div>
        {messages.counter.prefix}
        <span style={{ color: "#e60012" }}>{formatCount(uniqueFans)}</span>
        {messages.counter.suffix}
      </div>
      <div style={{ fontSize: 13 }}>
        <span style={{ color: "#e60012" }}>{formatCount(prefectureCount)}</span>
        {" / "}
        {PREFECTURES.length}
        {" "}
        {messages.counter.prefectureCoverageSuffix}
      </div>
    </div>
  );
}
