import { useEffect, useRef } from "react";
import maplibregl from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { fetchPins } from "./api";

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

  useEffect(() => {
    if (!containerRef.current) return;

    const map = new maplibregl.Map({
      container: containerRef.current,
      style,
      center: [137.5, 38.0],
      zoom: 4.3,
    });

    map.on("load", async () => {
      const res = await fetchPins();
      const geojson = {
        type: "FeatureCollection",
        features: res.pins.map((p) => ({
          type: "Feature",
          geometry: { type: "Point", coordinates: [p.lng, p.lat] },
          properties: { weight: p.weight ?? 1 },
        })),
      };
      map.addSource("pins", { type: "geojson", data: geojson as any });
      map.addLayer({
        id: "pins-heat",
        type: "heatmap",
        source: "pins",
        paint: {
          "heatmap-weight": ["coalesce", ["get", "weight"], 1],
          "heatmap-intensity": ["interpolate", ["linear"], ["zoom"], 4, 1, 9, 3],
          "heatmap-radius": ["interpolate", ["linear"], ["zoom"], 4, 18, 9, 40],
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
    });

    return () => map.remove();
  }, []);

  return <div id="map" ref={containerRef} />;
}
