// 読み込みエラー時の全幅バナー。再試行ボタンで map を作り直す。
export function ErrorBanner({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <div
      role="alert"
      style={{
        position: "absolute",
        top: 12,
        left: 12,
        right: 12,
        padding: "12px 16px",
        background: "rgba(166,71,26,0.95)",
        color: "#fff",
        borderRadius: 8,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        gap: 12,
        zIndex: 1,
      }}
    >
      <span>{message}</span>
      <button type="button" onClick={onRetry}>
        再試行
      </button>
    </div>
  );
}
