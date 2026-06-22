// 国土数値情報「行政区域データ(N03)」GeoJSON から、
//   - backend/internal/geo/data/municipalities.geojson（埋め込み用。代表点 rep_lng/rep_lat 付き）
//   - web/src/municipalities.ts（あいまい検索の実行時リスト）
// を生成する。バックエンドとフロントを同一の元データから作り、表記ゆれ・drift を防ぐ。
//
// 使い方（nix develop 内で `make gen-municipalities`、または直接）:
//   N03_GEOJSON=/path/to/N03-YY_GML/N03.geojson node generate.mjs
// 任意:
//   SIMPLIFY=8%        簡素化率（小さいほど軽く粗い。既定 8%）
//   KANA_CSV=/path.csv code,kana 列の読みかなマスタ（総務省 全国地方公共団体コード等）。あれば kana を補完
//
// N03 のプロパティ(2025年版): N03_001=都道府県名 / N03_003=郡名 / N03_004=市区町村名(政令市は市名) /
//   N03_005=政令市の区名 / N03_007=行政区域コード(5桁)。表示名は N03_003+N03_004+N03_005 を連結する。

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import mapshaper from "mapshaper";
import polylabel from "polylabel";

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, "..", "..");
const backendOut = path.join(repoRoot, "backend/internal/geo/data/municipalities.geojson");
const webOut = path.join(repoRoot, "web/src/municipalities.ts");

const src = process.env.N03_GEOJSON || process.argv[2];
if (!src) {
  console.error("N03_GEOJSON（または第1引数）で N03 GeoJSON のパスを指定してください");
  process.exit(1);
}
const simplify = process.env.SIMPLIFY || "8%";

// code -> kana の読みマスタ（任意）。
const kanaByCode = new Map();
if (process.env.KANA_CSV) {
  const csv = fs.readFileSync(process.env.KANA_CSV, "utf8").split(/\r?\n/);
  for (const line of csv) {
    const [code, kana] = line.split(",");
    if (code && kana) kanaByCode.set(code.trim(), kana.trim());
  }
}

// 1) mapshaper で簡素化し、市区町村コードで dissolve（同一コードの分割ポリゴンを統合）。
//    keep-shapes で簡素化による小島・細district の消失を防ぐ。
const cmd =
  `-i "${src}" ` +
  `-simplify ${simplify} keep-shapes ` +
  `-dissolve2 N03_007 copy-fields=N03_001,N03_003,N03_004,N03_005 ` +
  `-clean ` +
  `-o format=geojson precision=0.00001 out.json`;

// applyCommands は名前付き出力を戻り値の dict に格納する（"-"/stdout はこのバージョンでは
// 実 stdout に書かれ捕捉できないため、ファイル名を与えてメモリ上で受け取る）。
const output = await mapshaper.applyCommands(cmd, {});
const merged = JSON.parse(output["out.json"] ?? Object.values(output)[0]);

// 2) 代表内部点(polylabel)を計算。MultiPolygon は最大面積のポリゴンを採用する。
function ringArea(ring) {
  let a = 0;
  for (let i = 0, j = ring.length - 1; i < ring.length; j = i++) {
    a += (ring[j][0] + ring[i][0]) * (ring[j][1] - ring[i][1]);
  }
  return Math.abs(a / 2);
}
function repPoint(geom) {
  let best = null; // [polygonCoords, area]
  if (geom.type === "Polygon") best = [geom.coordinates, ringArea(geom.coordinates[0])];
  else if (geom.type === "MultiPolygon") {
    for (const poly of geom.coordinates) {
      const area = ringArea(poly[0]);
      if (!best || area > best[1]) best = [poly, area];
    }
  }
  if (!best) return null;
  const [lng, lat] = polylabel(best[0], 0.0005);
  return [lng, lat];
}

const features = [];
const muni = [];
for (const f of merged.features) {
  const p = f.properties ?? {};
  // N03_007 を厳密に5桁へ正規化する。数値化で先頭ゼロが落ちた入力(例: "1101")も
  // padStart で復元し、5桁にならないものは弾く（municipality_code/kana 結合のずれ防止）。
  let code = String(p.N03_007 ?? "").trim();
  if (/^\d{1,5}$/.test(code)) code = code.padStart(5, "0");
  if (!/^\d{5}$/.test(code) || !f.geometry) continue;
  const prefecture = p.N03_001 ?? "";
  // 郡名＋市区町村名＋政令市の区名を連結する（例: 石狩郡当別町 / 八王子市 / 札幌市中央区）。
  const name = `${p.N03_003 ?? ""}${p.N03_004 ?? ""}${p.N03_005 ?? ""}`.trim();
  if (!name) continue;
  const kana = kanaByCode.get(code) ?? "";
  const rep = repPoint(f.geometry);

  const props = { code, prefecture, name, kana };
  if (rep) {
    props.rep_lng = Number(rep[0].toFixed(6));
    props.rep_lat = Number(rep[1].toFixed(6));
  }
  features.push({ type: "Feature", properties: props, geometry: f.geometry });
  muni.push({ code, prefecture, name, kana });
}

// 重複コードがあれば最後勝ちで畳む（dissolve 済みだが念のため）。backend/frontend が
// 重複コードで食い違わないよう、features と muni を同じ基準で dedup・ソートする。
const featByCode = new Map(features.map((f) => [f.properties.code, f]));
const featSorted = [...featByCode.values()].sort((a, b) =>
  a.properties.code.localeCompare(b.properties.code),
);
const byCode = new Map(muni.map((m) => [m.code, m]));
const muniSorted = [...byCode.values()].sort((a, b) => a.code.localeCompare(b.code));

// 3) backend 埋め込み用 geojson を書き出す。
fs.writeFileSync(
  backendOut,
  JSON.stringify({ type: "FeatureCollection", features: featSorted }) + "\n",
);

// 4) frontend 用 TS を書き出す。
const tsBody = muniSorted
  .map(
    (m) =>
      `  { code: ${JSON.stringify(m.code)}, prefecture: ${JSON.stringify(m.prefecture)}, name: ${JSON.stringify(m.name)}, kana: ${JSON.stringify(m.kana)} },`,
  )
  .join("\n");
const ts = `import type { Prefecture } from "./prefectures";

// Municipality は投稿フォームの市区町村候補1件。
// code は全国地方公共団体コードで、投稿時に municipality_code として送る。
export type Municipality = {
  code: string;
  prefecture: Prefecture;
  name: string;
  kana: string;
};

// MUNICIPALITIES は市区町村の実行時リスト（あいまい検索の対象）。
// このファイルは tools/municipalities/generate.mjs の自動生成物。手で編集しないこと。
// backend/internal/geo/data/municipalities.geojson と同一の元データから生成している。
export const MUNICIPALITIES: readonly Municipality[] = [
${tsBody}
] as const;
`;
fs.writeFileSync(webOut, ts);

console.log(
  `生成完了: ${features.length}件\n  ${path.relative(repoRoot, backendOut)}\n  ${path.relative(repoRoot, webOut)}`,
);
