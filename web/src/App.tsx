import { useEffect, useRef, useState } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { fetchPins } from "./api";
import { logger } from "./logger";
import { messages } from "./messages";

const style: maplibregl.StyleSpecification = {
  version: 8,
  sources: {
    gsi: {
      type: "raster",
      tiles: ["https://cyberjapandata.gsi.go.jp/xyz/pale/{z}/{x}/{y}.png"],
      tileSize: 256,
      attribution:
        '<a href="https://maps.gsi.go.jp/development/ichiran.html">国土地理院</a>',
    },
  },
  layers: [{ id: "gsi", type: "raster", source: "gsi" }],
};

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
      style,
      center: [137.5, 38.0],
      zoom: 4.3,
    });

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
        map.addSource("pins", { type: "geojson", data: geojson });
        map.addLayer({
          id: "pins-heat",
          type: "heatmap",
          source: "pins",
          paint: {
            "heatmap-weight": ["coalesce", ["get", "weight"], 1],
            "heatmap-intensity": [
              "interpolate",
              ["linear"],
              ["zoom"],
              4,
              1,
              9,
              3,
            ],
            "heatmap-radius": [
              "interpolate",
              ["linear"],
              ["zoom"],
              4,
              18,
              9,
              40,
            ],
            "heatmap-color": [
              "interpolate",
              ["linear"],
              ["heatmap-density"],
              0,
              "rgba(217,123,58,0)",
              0.2,
              "rgba(240,217,198,0.6)",
              0.5,
              "rgb(217,123,58)",
              1,
              "rgb(166,71,26)",
            ],
            "heatmap-opacity": 0.85,
          },
        });
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
