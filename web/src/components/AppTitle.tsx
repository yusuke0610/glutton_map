import { messages } from "../lib/messages";

// アプリタイトル（上部中央）。背景は透過しフォントのみ。地図に埋もれないよう薄く影を付ける。
export function AppTitle() {
  return (
    <h1
      style={{
        position: "absolute",
        top: 12,
        left: "50%",
        transform: "translateX(-50%)",
        margin: 0,
        zIndex: 2,
        color: "#d97b3a",
        fontWeight: "bold",
        fontSize: 22,
        whiteSpace: "nowrap",
        textShadow: "0 1px 3px rgba(255,255,255,0.9)",
        pointerEvents: "none",
      }}
    >
      {messages.title}
    </h1>
  );
}
