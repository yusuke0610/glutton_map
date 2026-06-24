import { messages } from "../lib/messages";
import {
  buildShareUrl,
  buildTweetIntentUrl,
  defaultShareText,
} from "../lib/share";

// 投稿直後に出す X 共有導線。編集可能な文面 + UTM 付き共有 URL で intent を開く。
// apiBase は backend の公開 URL（共有リンクの /share がここを指す）。
const API_BASE = import.meta.env.VITE_API_BASE ?? "";

const cardStyle: React.CSSProperties = {
  position: "absolute",
  top: 16,
  left: "50%",
  transform: "translateX(-50%)",
  zIndex: 3,
  background: "rgba(255,255,255,0.97)",
  borderRadius: 10,
  padding: 16,
  boxShadow: "0 2px 12px rgba(0,0,0,0.2)",
  display: "flex",
  flexDirection: "column",
  gap: 10,
  width: 320,
  maxWidth: "90vw",
};

const xButtonStyle: React.CSSProperties = {
  background: "#000",
  color: "#fff",
  border: "none",
  borderRadius: 8,
  padding: "10px 14px",
  fontWeight: "bold",
  textDecoration: "none",
  textAlign: "center",
  cursor: "pointer",
};

const closeButtonStyle: React.CSSProperties = {
  background: "transparent",
  border: "none",
  color: "#666",
  fontSize: 20,
  lineHeight: 1,
  cursor: "pointer",
  padding: 0,
  width: 24,
  height: 24,
};

export function ShareButton({
  prefecture,
  onClose,
}: {
  prefecture?: string;
  onClose: () => void;
}) {
  const shareUrl = buildShareUrl(API_BASE, prefecture);
  const intentUrl = buildTweetIntentUrl(defaultShareText(prefecture), shareUrl);

  return (
    <div style={cardStyle}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <strong>{messages.share.heading}</strong>
        <button
          type="button"
          aria-label={messages.share.close}
          onClick={onClose}
          style={closeButtonStyle}
        >
          ×
        </button>
      </div>

      <a
        href={intentUrl}
        target="_blank"
        rel="noopener noreferrer"
        style={xButtonStyle}
      >
        {messages.share.button}
      </a>
    </div>
  );
}
