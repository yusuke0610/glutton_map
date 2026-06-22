import { DROP_TIMING } from "../map/pin-drop";

// ピン打ち込み演出の keyframes。ピンの落下と、着地時の弾み（squash & stretch）を定義する。
// 落下グループには translate(-50%,-100%) の基準位置があるため、transform にそれを含めて上書きしない。
const dropKeyframes = `
@keyframes pin-fall {
  from { transform: translate(-50%, calc(-100% - 220px)); }
  to { transform: translate(-50%, -100%); }
}
@keyframes pin-stick {
  0%, 80% { transform: scale(0.92, 1.12); }
  90% { transform: scale(1.08, 0.9); }
  100% { transform: scale(1, 1); }
}
`;

// 投稿時の打ち込み演出。地図上の画面座標 (x,y) にピンが上から落下→着地して刺さる。
export function PinDropOverlay({ x, y }: { x: number; y: number }) {
  return (
    <div
      aria-hidden
      style={{
        position: "absolute",
        left: x,
        top: y,
        pointerEvents: "none",
        zIndex: 10,
      }}
    >
      {/* keyframes。CSS ファイルを持たない方針のため style タグで注入する。 */}
      <style>{dropKeyframes}</style>
      {/* ピンを落下させるグループ。着地後は forwards で留まる。 */}
      <div
        style={{
          position: "absolute",
          // ピン先端が落下先 (left,top) に来るよう、下端中央を原点に合わせる。
          transform: "translate(-50%, -100%)",
          animation: `pin-fall ${DROP_TIMING.dropMs}ms cubic-bezier(0.5, 0, 0.9, 0.5) forwards`,
        }}
      >
        {/* ピン（プレースホルダ）。地図上の本番アイコン createPinIcon に合わせたティアドロップ。 */}
        <svg
          width={28}
          height={36}
          viewBox="0 0 28 36"
          style={{
            display: "block",
            transformOrigin: "bottom center",
            animation: `pin-stick ${DROP_TIMING.dropMs}ms ease-out forwards`,
          }}
        >
          <path
            d="M14 35 C5 22 1 16 1 10 A13 13 0 0 1 27 10 C27 16 23 22 14 35 Z"
            fill="#e60012"
            stroke="#fff"
            strokeWidth={2}
          />
          <circle cx={14} cy={11} r={5} fill="#fff" />
        </svg>
      </div>
    </div>
  );
}
