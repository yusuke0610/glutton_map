// コミット済みの市区町村境界（backend/internal/geo/data/municipalities.geojson）を
// 都道府県名(prefecture)で dissolve し、市区町村の内部境界を除いた47都道府県の外周ポリゴンを
//   - web/public/prefectures.geojson（フロントが実行時に読む静的アセット。赤線ハイライト用）
// として生成する。元データ(N03)に依存せず、コミット済みデータから作れる。
//
// 使い方（nix develop 内で `make gen-prefectures`、または直接）:
//   node generate-prefectures.mjs
// 任意:
//   SIMPLIFY=8%   簡素化率（小さいほど軽く粗い。ハイライト用途なので強めでよい。既定 8%）

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import mapshaper from "mapshaper";

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, "..", "..");
const src = path.join(repoRoot, "backend/internal/geo/data/municipalities.geojson");
const out = path.join(repoRoot, "web/public/prefectures.geojson");

const simplify = process.env.SIMPLIFY || "8%";

// dissolve2 prefecture で県内の市区町村を統合し内部境界を除去。
// filter-fields prefecture で各 feature のプロパティをフロントの filter キー(prefecture)のみに絞る。
// keep-shapes で簡素化による小島・離島の消失を防ぐ。
const cmd =
  `-i "${src}" ` +
  `-simplify ${simplify} keep-shapes ` +
  `-dissolve2 prefecture ` +
  `-filter-fields prefecture ` +
  `-clean ` +
  `-o format=geojson precision=0.0001 out.json`;

const output = await mapshaper.applyCommands(cmd, {});
const fc = JSON.parse(output["out.json"] ?? Object.values(output)[0]);

fs.mkdirSync(path.dirname(out), { recursive: true });
fs.writeFileSync(out, JSON.stringify(fc) + "\n");

console.log(`生成完了: ${fc.features.length}件\n  ${path.relative(repoRoot, out)}`);
