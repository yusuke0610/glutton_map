import { useCallback, useEffect, useRef, useState } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { fetchPins, type Pin } from "./api/api";
import { shouldAnimateDrop, flyToOptionsFor, DROP_TIMING } from "./map/pin-drop";
import { logger } from "./lib/logger";
import { messages } from "./lib/messages";
import { mapStyle } from "./map/mapStyle";
import { popupHTML } from "./map/popup";
import { createPinIcon } from "./map/pinIcon";
import { buildPinGeojson } from "./map/pinGeojson";
import {
  heatmapLayer,
  PIN_ICON_IMAGE,
  PIN_ICON_LAYER_ID,
  PINS_SOURCE_ID,
  pinIconLayer,
} from "./map/pinLayers";
import { AppTitle } from "./components/AppTitle";
import { HeroCounter } from "./components/HeroCounter";
import { PinDropOverlay } from "./components/PinDropOverlay";
import { PinForm } from "./components/PinForm";
import { ErrorBanner } from "./components/ErrorBanner";

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

  return (
    <div style={{ position: "relative", height: "100%" }}>
      <div id="map" ref={containerRef} />

      {drop && <PinDropOverlay x={drop.x} y={drop.y} />}

      <AppTitle />

      {total !== null && <HeroCounter total={total} />}

      {/* 投稿フォーム（右上）。読み込みエラー時は全幅バナーと被るため非表示にする。 */}
      <PinForm hidden={!!error} onSubmitted={playDropAndZoom} />

      {error && <ErrorBanner message={error} onRetry={handleRetry} />}
    </div>
  );
}
