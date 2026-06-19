# 如月愛マップ

日本地図（国土地理院 淡色タイル）上に、API から取得したピンをヒートマップ描画する縦割り1スライス。

契約は `backend/openapi.yaml` が唯一の真実（spec-first）。Go サーバ型は oapi-codegen、
フロント TS 型は openapi-typescript で同じ yaml から生成する（両側手書き禁止）。

## 構成
- `backend/` Go 1.22 + Gin + oapi-codegen（strict-server）+ modernc.org/sqlite（pure Go）
- `web/` Vite + React 18 + TypeScript + MapLibre GL JS

## 実行手順

開発タスク（型生成・ビルド・コンテナ起動・テスト・Lint 等）はすべて **`Makefile` に集約**している。
`nix develop` のシェル内で `make help` を実行すると、利用可能なターゲットの一覧が出る。
個別のコマンドは README には書かず、`Makefile` を唯一の入口とする。

## ヒーロー指標
`prefecture_count` =「何都道府県に散らばっているか」。人数より広がりを主役にする。
