import { counterText, formatCount } from "../lib/counter";
import { messages } from "../lib/messages";

// ヒーロー指標（左上）。数字部分だけ赤で強調する（例: 全世界にくいしんぼが◯◯人！）。
export function HeroCounter({ total }: { total: number }) {
  return (
    <div
      aria-label={counterText(total)}
      style={{
        position: "absolute",
        top: 12,
        left: 12,
        zIndex: 2,
        padding: "8px 14px",
        background: "rgba(255,255,255,0.95)",
        color: "#d97b3a",
        borderRadius: 999,
        fontWeight: "bold",
        fontSize: 16,
        boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
        pointerEvents: "none",
      }}
    >
      {messages.counter.prefix}
      <span style={{ color: "#e60012" }}>{formatCount(total)}</span>
      {messages.counter.suffix}
    </div>
  );
}
