# 市区町村データ生成（行政区域 → 境界データ）

投稿フォームの市区町村候補と、ピンを「境界の内側」に立てるためのポリゴンを、
**国土数値情報「行政区域データ(N03)」** から生成する。出力は2つ:

- `backend/internal/geo/data/municipalities.geojson` … バックエンドが `//go:embed` する境界データ（代表点 `rep_lng`/`rep_lat` 付き）
- `web/src/municipalities.ts` … フロントのあいまい検索の実行時リスト（自動生成・手編集禁止）

両方を同一の元データから生成し、表記ゆれ・drift を防ぐ。

## 手順

1. 元データを入手する（手動DL）。
   - 国土数値情報ダウンロードサイトの「行政区域データ(N03)」から全国版を取得し、Shapefile→GeoJSON 変換するか、GeoJSON 配布を使う。
   - クレジット表示（出典: 国土数値情報）が利用条件。アプリ/README に明記すること。
2. （任意・読み検索を効かせる場合）総務省「全国地方公共団体コード」から KANA_CSV を作る。
   - 総務省CSVは複数行ヘッダ・6桁コード・半角カナで generate.mjs に直接渡せないため、`prep-kana.mjs` で `code,kana` に整形する。
   - 政令市の区は全国版CSVに無く別ファイル（政令指定都市の区版）に入るので、両方をカンマ区切りで渡す（後勝ちでマージ）。
   ```bash
   make gen-kana KANA_SRC=/path/全国版.csv,/path/政令市区版.csv
   # 既定で backend/data/kana_clean.csv に出力（KANA_OUT で変更可）
   ```
3. 生成する（`nix develop` 内）:
   ```bash
   make gen-municipalities N03_GEOJSON=/path/to/N03.geojson KANA_CSV=/path/to/kana_clean.csv
   # 任意: SIMPLIFY=8%（簡素化率）。KANA_CSV 省略時は漢字の部分一致のみ
   ```
4. 生成物を `git add` して型・テストを通す:
   ```bash
   make gen        # 念のため型再生成
   make test       # backend + web
   ```

## メモ

- `kana` は N03 に含まれないため、読みでの検索を効かせたい場合は総務省CSVから `prep-kana.mjs`（`make gen-kana`）で `KANA_CSV` を作って渡す。未指定なら漢字の部分一致のみ。`kana_clean.csv` は中間生成物（gitignore）で、元の総務省CSVを消しても再DL→`make gen-kana` で再現できる。
- 代表点は polylabel（最大面積ポリゴンの最深内部点）で事前計算し、バックエンドのロードを高速化する。バックエンドは代表点が境界外のときだけグリッド走査でフォールバックする。
- 簡素化を強めすぎると境界が粗くなり「隣の区に入る」可能性が上がる。`keep-shapes` で小島の消失を防ぎつつ `SIMPLIFY` を調整する。
