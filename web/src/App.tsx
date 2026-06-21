import { useEffect, useRef, useState } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { fetchPins } from "./api";
import { logger } from "./logger";
import { messages } from "./messages";
import { mapStyle } from "./mapStyle";
import {
  heatmapLayer,
  PIN_ICON_IMAGE,
  PINS_SOURCE_ID,
  pinIconLayer,
} from "./pinLayers";

// ピン上部に載せるアイコンを生成する（マーカー型ピン＋上部に顔アイコン）。
// 顔の部分は今はプレースホルダ。公式の食いしんboy アイコンの許可が下りたら、
// 下の「公式アイコンに差し替える箇所」で画像を drawImage するだけで本番化できる。
function createPinIcon(): { image: ImageData; options: { pixelRatio: number } } {
  const w = 40;
  const h = 52;
  const ratio = 2;
  const canvas = document.createElement("canvas");
  canvas.width = w * ratio;
  canvas.height = h * ratio;
  const ctx = canvas.getContext("2d")!;
  ctx.scale(ratio, ratio);

  const color = "#d97b3a"; // ヒートマップと揃えた暖色
  // 下に伸びる尖り（先端が座標を指す）。
  ctx.beginPath();
  ctx.moveTo(9, 27);
  ctx.lineTo(20, 50);
  ctx.lineTo(31, 27);
  ctx.closePath();
  ctx.fillStyle = color;
  ctx.fill();
  // 頭（丸）。
  ctx.beginPath();
  ctx.arc(20, 18, 16, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
  ctx.lineWidth = 2;
  ctx.strokeStyle = "#ffffff";
  ctx.stroke();

  // --- 公式アイコンに差し替える箇所（今は「如」の文字）---
  ctx.fillStyle = "#ffffff";
  ctx.font = "bold 20px sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText("如", 20, 19);
  // --- ここまで ---

  return {
    image: ctx.getImageData(0, 0, canvas.width, canvas.height),
    options: { pixelRatio: ratio },
  };
}

export default function App() {
  const containerRef = useRef<HTMLDivElement>(null);
  // ユーザー向けエラー文言（null = エラーなし）。
  const [error, setError] = useState<string | null>(null);
  // 再試行トリガ。値を増やすと useEffect が再実行され map を作り直す。
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    if (!containerRef.current) return;

    const map = new maplibregl.Map({
      container: containerRef.current,
      style: mapStyle,
      center: [137.5, 38.0],
      zoom: 4.3,
      // ズームアウトしすぎて地図が極小になるのを防ぐ（Google マップ程度の下限）。
      minZoom: 3,
      // 北を常に上に固定する（上=ロシア / 下=南極の向きを保つ）。
      // ドラッグでの回転（bearing）と傾き（pitch）を無効化する。
      dragRotate: false, // 右ドラッグ/Ctrl+ドラッグでの回転・チルト
      pitchWithRotate: false,
      touchPitch: false, // 2本指でのチルト
      maxPitch: 0, // 傾きを 0 に固定
    });

    // タッチ操作の回転、キーボードでの回転も無効化する（ピンチズームは残す）。
    map.touchZoomRotate.disableRotation();
    map.keyboard.disableRotation();

    // E2E から地図状態（bearing/pitch 等）を検証できるよう公開する。
    (window as unknown as { __map: maplibregl.Map }).__map = map;

    map.on("load", async () => {
      try {
        const res = await fetchPins();
        const geojson: GeoJSON.FeatureCollection = {
          type: "FeatureCollection",
          features: res.pins.map((p) => ({
            type: "Feature",
            geometry: { type: "Point", coordinates: [p.lng, p.lat] },
            properties: { weight: p.weight ?? 1 },
          })),
        };
        // ピンアイコン（プレースホルダ）を登録。
        if (!map.hasImage(PIN_ICON_IMAGE)) {
          const icon = createPinIcon();
          map.addImage(PIN_ICON_IMAGE, icon.image, icon.options);
        }
        map.addSource(PINS_SOURCE_ID, { type: "geojson", data: geojson });
        // ハイブリッド表示: ズームアウト=ヒートマップ（分布）、ズームイン=ピン（個別）。
        map.addLayer(heatmapLayer());
        map.addLayer(pinIconLayer());
        setError(null);
      } catch (e) {
        // 開発者向けはフロー追従で発生箇所のログに、ユーザー向けは一元管理の文言を表示。
        logger.error("地図データの読み込みに失敗", e);
        setError(messages.error.fetchPins);
      }
    });

    return () => map.remove();
  }, [reloadKey]);

  // 再試行: キーを増やして useEffect を再実行し、map を作り直す。
  const handleRetry = () => {
    setError(null);
    setReloadKey((k) => k + 1);
  };

  return (
    <div style={{ position: "relative", height: "100%" }}>
      <div id="map" ref={containerRef} />
      {error && (
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
          <span>{error}</span>
          <button type="button" onClick={handleRetry}>
            再試行
          </button>
        </div>
      )}
    </div>
  );
}
