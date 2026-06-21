import { useEffect, useRef, useState } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { createPin, fetchPins } from "./api";
import { counterText, formatCount } from "./counter";
import { logger } from "./logger";
import { messages } from "./messages";
import { mapStyle } from "./mapStyle";
import { PREFECTURES, type Prefecture } from "./prefectures";
import { popupHTML } from "./popup";
import {
  heatmapLayer,
  PIN_ICON_IMAGE,
  PIN_ICON_LAYER_ID,
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

// 投稿フォームの共通スタイル。
const panelButtonStyle: React.CSSProperties = {
  background: "#d97b3a",
  color: "#fff",
  border: "none",
  borderRadius: 8,
  padding: "10px 14px",
  fontWeight: "bold",
  cursor: "pointer",
};
const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 4,
  fontSize: 13,
  color: "#333",
};
const inputStyle: React.CSSProperties = {
  padding: "6px 8px",
  borderRadius: 6,
  border: "1px solid #ccc",
  font: "inherit",
};
// 罰（×）の閉じるボタン。テキストではなくアイコン表示にする。
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

export default function App() {
  const containerRef = useRef<HTMLDivElement>(null);
  // ユーザー向けエラー文言（null = エラーなし）。
  const [error, setError] = useState<string | null>(null);
  // 再試行トリガ。値を増やすと useEffect が再実行され map を作り直す。
  const [reloadKey, setReloadKey] = useState(0);
  // ピン総数（左上のヒーロー表示用）。null = 未取得。
  const [total, setTotal] = useState<number | null>(null);

  // 投稿フォームの状態。
  const [formOpen, setFormOpen] = useState(false);
  const [nickname, setNickname] = useState("");
  const [prefecture, setPrefecture] = useState<Prefecture | "">("");
  const [city, setCity] = useState("");
  const [comment, setComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  // 投稿結果の通知（成功 or 失敗）。
  const [formNotice, setFormNotice] = useState<
    { kind: "success" | "error"; text: string } | null
  >(null);

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
            properties: {
              weight: p.weight ?? 1,
              // ポップアップ表示用。seed 由来のピンは空文字。
              prefecture: p.prefecture ?? "",
              nickname: p.nickname ?? "",
              city: p.city ?? "",
              comment: p.comment ?? "",
            },
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
        setTotal(res.total);
        setError(null);
      } catch (e) {
        // 開発者向けはフロー追従で発生箇所のログに、ユーザー向けは一元管理の文言を表示。
        logger.error("地図データの読み込みに失敗", e);
        setError(messages.error.fetchPins);
      }
    });

    // ピンをクリックすると投稿内容をポップアップ表示する（ズームイン時のみピンが出る）。
    const popup = new maplibregl.Popup({ closeButton: true, closeOnClick: true });
    map.on("click", PIN_ICON_LAYER_ID, (e) => {
      const f = e.features?.[0];
      if (!f) return;
      const props = f.properties ?? {};
      const [lng, lat] = (f.geometry as GeoJSON.Point).coordinates;
      popup
        .setLngLat([lng, lat])
        .setHTML(
          popupHTML({
            nickname: props.nickname || undefined,
            prefecture: props.prefecture || undefined,
            city: props.city || undefined,
            comment: props.comment || undefined,
          }),
        )
        .addTo(map);
    });
    map.on("mouseenter", PIN_ICON_LAYER_ID, () => {
      map.getCanvas().style.cursor = "pointer";
    });
    map.on("mouseleave", PIN_ICON_LAYER_ID, () => {
      map.getCanvas().style.cursor = "";
    });

    return () => map.remove();
  }, [reloadKey]);

  // 再試行: キーを増やして useEffect を再実行し、map を作り直す。
  const handleRetry = () => {
    setError(null);
    setReloadKey((k) => k + 1);
  };

  // 投稿: createPin で送信し、成功したらマップを再取得して新しいピンを反映する。
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!prefecture || submitting) return;
    setSubmitting(true);
    setFormNotice(null);
    try {
      await createPin({
        nickname,
        prefecture,
        city,
        comment: comment || undefined,
      });
      setFormNotice({ kind: "success", text: messages.form.success });
      // 入力をリセットし、マップを作り直して投稿を反映する。
      setNickname("");
      setPrefecture("");
      setCity("");
      setComment("");
      setReloadKey((k) => k + 1);
    } catch (err) {
      logger.error("ピンの投稿に失敗", err);
      setFormNotice({ kind: "error", text: messages.error.createPin });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div style={{ position: "relative", height: "100%" }}>
      <div id="map" ref={containerRef} />

      {/* ヒーロー指標（左上）。数字部分だけ赤で強調する（例: 全世界にくいしんぼが◯◯人！）。 */}
      {total !== null && (
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
      )}

      {/* 投稿フォーム（右上のパネル）。閉じている間はトグルボタンのみ。
          読み込みエラー時は全幅のエラーバナーと被るため非表示にする。 */}
      {!error && (
      <div
        style={{
          position: "absolute",
          top: 16,
          right: 16,
          zIndex: 2,
          width: formOpen ? 280 : "auto",
        }}
      >
        {!formOpen ? (
          <button
            type="button"
            onClick={() => setFormOpen(true)}
            style={panelButtonStyle}
          >
            {messages.form.open}
          </button>
        ) : (
          <form
            onSubmit={handleSubmit}
            aria-label={messages.form.title}
            style={{
              background: "rgba(255,255,255,0.97)",
              borderRadius: 10,
              padding: 16,
              boxShadow: "0 2px 12px rgba(0,0,0,0.2)",
              display: "flex",
              flexDirection: "column",
              gap: 10,
            }}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <strong>{messages.form.title}</strong>
              <button
                type="button"
                aria-label={messages.form.close}
                onClick={() => setFormOpen(false)}
                style={closeButtonStyle}
              >
                ×
              </button>
            </div>

            <label style={labelStyle}>
              {messages.form.nickname}
              <input
                type="text"
                required
                maxLength={30}
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                style={inputStyle}
              />
            </label>

            <label style={labelStyle}>
              {messages.form.prefecture}
              <select
                required
                value={prefecture}
                onChange={(e) => setPrefecture(e.target.value as Prefecture)}
                style={inputStyle}
              >
                <option value="" disabled>
                  ―
                </option>
                {PREFECTURES.map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              {messages.form.city}
              <input
                type="text"
                required
                maxLength={50}
                value={city}
                onChange={(e) => setCity(e.target.value)}
                style={inputStyle}
              />
            </label>

            <label style={labelStyle}>
              {messages.form.comment}
              <textarea
                maxLength={200}
                rows={3}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                style={inputStyle}
              />
            </label>

            <button type="submit" disabled={submitting} style={panelButtonStyle}>
              {submitting ? messages.form.submitting : messages.form.submit}
            </button>

            {formNotice && (
              <div
                role={formNotice.kind === "error" ? "alert" : "status"}
                style={{
                  color: formNotice.kind === "error" ? "#a6471a" : "#1a7a3a",
                  fontSize: 13,
                }}
              >
                {formNotice.text}
              </div>
            )}
          </form>
        )}
      </div>
      )}

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
