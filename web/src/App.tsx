import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { createPin, fetchPins, type Pin } from "./api";
import { shouldAnimateDrop, flyToOptionsFor, DROP_TIMING } from "./pin-drop";
import { counterText, formatCount } from "./counter";
import { logger } from "./logger";
import { messages } from "./messages";
import { mapStyle } from "./mapStyle";
import { PREFECTURES, type Prefecture } from "./prefectures";
import { MUNICIPALITIES } from "./municipalities";
import { searchMunicipalities } from "./municipality-search";
import { canSubmitPin } from "./pin-form";
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

// ピン配列を地図ソース用の GeoJSON に変換する。読み込み時と投稿後の更新で共用する。
function buildPinGeojson(pins: Pin[]): GeoJSON.FeatureCollection {
  return {
    type: "FeatureCollection",
    features: pins.map((p) => ({
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
}

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
// 無効時（市区町村を候補から選んでいない等）は白くぼかして押せないことを伝える。
const panelButtonDisabledStyle: React.CSSProperties = {
  background: "#e8e2dc",
  color: "#fff",
  cursor: "not-allowed",
  opacity: 0.6,
  filter: "blur(0.4px)",
  boxShadow: "none",
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
// 市区町村のあいまい検索候補リスト（入力欄の下に重ねて表示）。
const suggestionListStyle: React.CSSProperties = {
  position: "absolute",
  top: "100%",
  left: 0,
  right: 0,
  margin: "2px 0 0",
  padding: 0,
  listStyle: "none",
  background: "#fff",
  border: "1px solid #ccc",
  borderRadius: 6,
  boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
  maxHeight: 180,
  overflowY: "auto",
  zIndex: 3,
};
const suggestionItemStyle: React.CSSProperties = {
  display: "block",
  width: "100%",
  textAlign: "left",
  background: "transparent",
  border: "none",
  padding: "8px 10px",
  font: "inherit",
  cursor: "pointer",
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
  // 地図インスタンスとポップアップを保持し、投稿後の flyTo / popup から参照する。
  const mapRef = useRef<maplibregl.Map | null>(null);
  const popupRef = useRef<maplibregl.Popup | null>(null);
  // ユーザー向けエラー文言（null = エラーなし）。
  const [error, setError] = useState<string | null>(null);
  // 再試行トリガ。値を増やすと useEffect が再実行され map を作り直す。
  const [reloadKey, setReloadKey] = useState(0);
  // ピン総数（左上のヒーロー表示用）。null = 未取得。
  const [total, setTotal] = useState<number | null>(null);
  // ピン打ち込み演出。値があるとき地図上の画面座標 (x,y) に手＋ピンを描画する。
  const [drop, setDrop] = useState<{ x: number; y: number } | null>(null);

  // 投稿フォームの状態。
  const [formOpen, setFormOpen] = useState(false);
  const [nickname, setNickname] = useState("");
  const [prefecture, setPrefecture] = useState<Prefecture | "">("");
  const [city, setCity] = useState("");
  // 選択された市区町村の全国地方公共団体コード。空 = 未選択（自由入力のフォールバック）。
  const [municipalityCode, setMunicipalityCode] = useState("");
  // 市区町村入力のフォーカス状態（候補リストの表示制御）。
  const [cityFocused, setCityFocused] = useState(false);
  const [comment, setComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  // 投稿結果の通知（成功 or 失敗）。
  const [formNotice, setFormNotice] = useState<
    { kind: "success" | "error"; text: string } | null
  >(null);

  // 市区町村のあいまい検索候補。コード選択済み（municipalityCode!=""）のときは出さない。
  const citySuggestions = useMemo(
    () =>
      municipalityCode === ""
        ? searchMunicipalities(MUNICIPALITIES, prefecture, city, 8)
        : [],
    [prefecture, city, municipalityCode],
  );

  // ピンを取得して地図ソースへ反映する。初回はソース／レイヤーを追加し、以降は setData で更新する。
  // 読み込み時（map.on("load")）と投稿成功後の両方から呼ぶ。
  const refreshPins = useCallback(async (map: maplibregl.Map) => {
    const res = await fetchPins();
    const geojson = buildPinGeojson(res.pins);
    const source = map.getSource(PINS_SOURCE_ID) as
      | maplibregl.GeoJSONSource
      | undefined;
    if (source) {
      source.setData(geojson);
    } else {
      // ピンアイコン（プレースホルダ）を登録。
      if (!map.hasImage(PIN_ICON_IMAGE)) {
        const icon = createPinIcon();
        map.addImage(PIN_ICON_IMAGE, icon.image, icon.options);
      }
      map.addSource(PINS_SOURCE_ID, { type: "geojson", data: geojson });
      // ハイブリッド表示: ズームアウト=ヒートマップ（分布）、ズームイン=ピン（個別）。
      map.addLayer(heatmapLayer());
      map.addLayer(pinIconLayer());
    }
    setTotal(res.total);
    setError(null);
  }, []);

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
    mapRef.current = map;

    map.on("load", async () => {
      try {
        await refreshPins(map);
      } catch (e) {
        // 開発者向けはフロー追従で発生箇所のログに、ユーザー向けは一元管理の文言を表示。
        logger.error("地図データの読み込みに失敗", e);
        setTotal(null);
        setError(messages.error.fetchPins);
      }
    });

    // ピンをクリックすると投稿内容をポップアップ表示する（ズームイン時のみピンが出る）。
    const popup = new maplibregl.Popup({ closeButton: true, closeOnClick: true });
    popupRef.current = popup;
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

    return () => {
      map.remove();
      mapRef.current = null;
      popupRef.current = null;
    };
  }, [reloadKey, refreshPins]);

  // 再試行: キーを増やして useEffect を再実行し、map を作り直す。
  const handleRetry = () => {
    setError(null);
    setReloadKey((k) => k + 1);
  };

  // 投稿したピンの位置に popup を表示する（投稿直後に自分の投稿を見せる）。
  const showPinPopup = (map: maplibregl.Map, pin: Pin) => {
    popupRef.current
      ?.setLngLat([pin.lng, pin.lat])
      .setHTML(
        popupHTML({
          nickname: pin.nickname || undefined,
          prefecture: pin.prefecture || undefined,
          city: pin.city || undefined,
          comment: pin.comment || undefined,
        }),
      )
      .addTo(map);
  };

  // 投稿成功後の演出: ピンを上から落として刺し → 着地でピンを反映 → 市区町村へズーム → popup 表示。
  // prefers-reduced-motion のときは落下演出を省き、ズームと popup だけ実行する。
  const playDropAndZoom = (pin: Pin) => {
    const map = mapRef.current;
    if (!map) return;
    const opts = flyToOptionsFor(pin.lng, pin.lat);
    const zoomThenPopup = () => {
      map.flyTo(opts);
      map.once("moveend", () => showPinPopup(map, pin));
    };
    const prefersReducedMotion =
      typeof window !== "undefined" &&
      window.matchMedia?.("(prefers-reduced-motion: reduce)").matches === true;

    if (!shouldAnimateDrop(prefersReducedMotion)) {
      void refreshPins(map).catch((e) => logger.error("ピン反映に失敗", e));
      zoomThenPopup();
      return;
    }

    // 現在の画面座標へ手＋ピンを落とす。
    const pt = map.project([pin.lng, pin.lat]);
    setDrop({ x: pt.x, y: pt.y });
    // 着地と同時にピンを地図へ反映する。
    window.setTimeout(() => {
      void refreshPins(map).catch((e) => logger.error("ピン反映に失敗", e));
    }, DROP_TIMING.impactMs);
    // 着地後に一拍おいてから演出を片付け、カメラを寄せて popup を出す。
    window.setTimeout(() => {
      setDrop(null);
      zoomThenPopup();
    }, DROP_TIMING.dropMs + DROP_TIMING.settleMs);
  };

  // 投稿: createPin で送信し、成功したら打ち込み演出 → 投稿地点へズームインする。
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // 市区町村は候補から選択（municipalityCode あり）されていないと投稿できない。
    // canSubmitPin は prefecture !== "" も確認するが、tsc に prefecture を絞り込ませる
    // ため明示の早期 return も置く（これがないと createPin に "" が渡りうると判定される）。
    if (!canSubmitPin({ prefecture, municipalityCode, submitting }) || prefecture === "")
      return;
    setSubmitting(true);
    setFormNotice(null);
    try {
      const created = await createPin({
        nickname,
        prefecture,
        city,
        municipality_code: municipalityCode,
        comment: comment || undefined,
      });
      setFormNotice({ kind: "success", text: messages.form.success });
      // 入力をリセットし、打ち込み演出とともに投稿を地図へ反映する。
      setNickname("");
      setPrefecture("");
      setCity("");
      setMunicipalityCode("");
      setComment("");
      playDropAndZoom(created);
    } catch (err) {
      logger.error("ピンの投稿に失敗", err);
      setFormNotice({ kind: "error", text: messages.error.createPin });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div style={{ position: "relative", height: "100%" }}>
      {/* ピン打ち込み演出の keyframes。CSS ファイルを持たない方針のため style タグで注入する。 */}
      <style>{dropKeyframes}</style>
      <div id="map" ref={containerRef} />

      {/* 投稿時の打ち込み演出。ピンが上から落下→着地して刺さる。 */}
      {drop && (
        <div
          aria-hidden
          style={{
            position: "absolute",
            left: drop.x,
            top: drop.y,
            pointerEvents: "none",
            zIndex: 10,
          }}
        >
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
      )}

      {/* アプリタイトル（上部中央）。背景は透過しフォントのみ。地図に埋もれないよう薄く影を付ける。 */}
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
                onChange={(e) => {
                  setPrefecture(e.target.value as Prefecture);
                  // 都道府県を変えたら市区町村の選択をリセットする。
                  setCity("");
                  setMunicipalityCode("");
                }}
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
              {/* 入力欄を他フォームと同じ幅に揃えるため flex で input を伸ばす。 */}
              <div
                style={{
                  position: "relative",
                  display: "flex",
                  flexDirection: "column",
                }}
              >
                <input
                  type="text"
                  required
                  maxLength={50}
                  value={city}
                  role="combobox"
                  aria-expanded={cityFocused && citySuggestions.length > 0}
                  aria-autocomplete="list"
                  autoComplete="off"
                  onChange={(e) => {
                    // 手入力したらコード選択を解除（候補から選び直す or 自由入力フォールバック）。
                    setCity(e.target.value);
                    setMunicipalityCode("");
                  }}
                  onFocus={() => setCityFocused(true)}
                  // クリック確定（onMouseDown）を取りこぼさないよう遅延して閉じる。
                  onBlur={() => setTimeout(() => setCityFocused(false), 120)}
                  style={inputStyle}
                />
                {cityFocused && citySuggestions.length > 0 && (
                  <ul role="listbox" style={suggestionListStyle}>
                    {citySuggestions.map((m) => (
                      <li key={m.code} role="option" aria-selected={false}>
                        <button
                          type="button"
                          // onMouseDown は input の onBlur より先に発火するので選択が確実に通る。
                          onMouseDown={(e) => {
                            e.preventDefault();
                            setCity(m.name);
                            setMunicipalityCode(m.code);
                            setCityFocused(false);
                          }}
                          // キーボード操作(Enter/Space)は click で発火するため、code 未設定で
                          // 都道府県フォールバックに落ちないよう同じ選択処理を割り当てる。
                          onClick={() => {
                            setCity(m.name);
                            setMunicipalityCode(m.code);
                            setCityFocused(false);
                          }}
                          style={suggestionItemStyle}
                        >
                          {m.name}
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
              {city !== "" && municipalityCode === "" && (
                <span role="alert" style={{ color: "#a6471a", fontSize: 12 }}>
                  {messages.form.cityHint}
                </span>
              )}
            </label>

            <label style={labelStyle}>
              {messages.form.comment}
              <textarea
                maxLength={200}
                rows={3}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                // 縦のみリサイズ可とし、パディング込みでもパネル幅を超えないようにする。
                style={{
                  ...inputStyle,
                  resize: "vertical",
                  boxSizing: "border-box",
                  width: "100%",
                  maxWidth: "100%",
                }}
              />
            </label>

            <button
              type="submit"
              disabled={!canSubmitPin({ prefecture, municipalityCode, submitting })}
              style={
                canSubmitPin({ prefecture, municipalityCode, submitting })
                  ? panelButtonStyle
                  : { ...panelButtonStyle, ...panelButtonDisabledStyle }
              }
            >
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
