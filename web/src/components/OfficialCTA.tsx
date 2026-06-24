import { messages } from "../lib/messages";
import { OUTBOUND, outboundUrl } from "../lib/outbound";

// 常設の公式送客 CTA（画面左下）。注文・メニューへ近い導線を優先する。
// リンクは backend の /out 経由でクリックを計測してから公式 URL へ 302 する。
const API_BASE = import.meta.env.VITE_API_BASE ?? "";

const ctaStyle: React.CSSProperties = {
  position: "absolute",
  left: 16,
  bottom: 16,
  zIndex: 2,
  background: "#d97b3a",
  color: "#fff",
  borderRadius: 8,
  padding: "10px 14px",
  fontWeight: "bold",
  textDecoration: "none",
  boxShadow: "0 2px 8px rgba(0,0,0,0.2)",
};

export function OfficialCTA() {
  return (
    <a
      href={outboundUrl(API_BASE, OUTBOUND.menu)}
      target="_blank"
      rel="noopener noreferrer"
      style={ctaStyle}
    >
      {messages.official.cta}
    </a>
  );
}
