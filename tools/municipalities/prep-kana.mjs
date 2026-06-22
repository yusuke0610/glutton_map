// 総務省「全国地方公共団体コード」CSV から、generate.mjs に渡す KANA_CSV
// (= code,kana のクリーンCSV) を生成する。
//
// 総務省CSVは複数行ヘッダ・UTF-8 BOM・6桁コード・半角カナを含み、generate.mjs の
// 素朴なパーサ(line.split(","))では読めないため、ここで前処理する。
//   - 6桁団体コードの先頭5桁(N03_007 と一致)を code にする
//   - 市区町村名カナ(5列目)を NFKC で全角カタカナ化して kana にする
//   - 政令市の区は全国版に無く別ファイル(政令指定都市の区版)に入るため、複数CSVを
//     順にマージする(後勝ち)。全国版→政令市区版 の順で渡すと区コードが追加される。
//
// 使い方(nix develop 内で `make gen-kana`、または直接):
//   KANA_SRC=/path/全国版.csv,/path/政令市区版.csv node prep-kana.mjs
// 任意:
//   KANA_OUT=/path/out.csv   出力先(既定 backend/data/kana_clean.csv)

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, "..", "..");

const srcEnv = process.env.KANA_SRC || process.argv[2];
if (!srcEnv) {
  console.error(
    "KANA_SRC(カンマ区切り)で総務省CSVを指定してください。例: 全国版.csv,政令市区版.csv",
  );
  process.exit(1);
}
const srcs = srcEnv.split(",").map((s) => s.trim()).filter((s) => s.length > 0);
const out = process.env.KANA_OUT || path.join(repoRoot, "backend/data/kana_clean.csv");

// ダブルクオート内の改行・カンマに対応した CSV パーサ。
function parseCSV(text) {
  if (text.charCodeAt(0) === 0xfeff) text = text.slice(1); // BOM 除去
  const rows = [];
  let row = [];
  let cur = "";
  let q = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (q) {
      if (c === '"') {
        if (text[i + 1] === '"') {
          cur += '"';
          i++;
        } else q = false;
      } else cur += c;
    } else if (c === '"') q = true;
    else if (c === ",") {
      row.push(cur);
      cur = "";
    } else if (c === "\r") {
      // CRLF の CR は無視
    } else if (c === "\n") {
      row.push(cur);
      rows.push(row);
      row = [];
      cur = "";
    } else cur += c;
  }
  if (cur.length > 0 || row.length > 0) {
    row.push(cur);
    rows.push(row);
  }
  return rows;
}

// 5桁コード -> カナ。複数CSVを順にマージ(後勝ち)。
const byCode = new Map();
for (const src of srcs) {
  const rows = parseCSV(fs.readFileSync(src, "utf8"))
    .slice(1) // ヘッダ行を捨てる
    .filter((r) => r[0] && /^[0-9]+$/.test(r[0]));
  let n = 0;
  for (const r of rows) {
    const code = r[0].slice(0, 5);
    const kana = (r[4] ?? "").normalize("NFKC").trim();
    if (kana.length > 0) {
      byCode.set(code, kana);
      n++;
    }
  }
  console.log(`${path.relative(repoRoot, src)}: ${n}件`);
}

const lines = [...byCode.entries()].map(([c, k]) => `${c},${k}`).join("\n") + "\n";
fs.writeFileSync(out, lines);
console.log(`生成完了: ${byCode.size}件 -> ${path.relative(repoRoot, out)}`);

// 生成済み municipalities.geojson があれば網羅率も出す(任意の参考情報)。
const geoPath = path.join(repoRoot, "backend/internal/geo/data/municipalities.geojson");
if (fs.existsSync(geoPath)) {
  const geo = JSON.parse(fs.readFileSync(geoPath, "utf8"));
  const codes = geo.features.map((f) => f.properties.code);
  const miss = codes.filter((c) => byCode.has(c) === false);
  console.log(`網羅: ${codes.length - miss.length}/${codes.length}件にカナあり(欠落 ${miss.length})`);
}
