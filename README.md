# 如月愛マップ

日本地図（国土地理院 淡色タイル）上に、API から取得したピンをヒートマップ描画する縦割り1スライス。

契約は `backend/openapi.yaml` が唯一の真実（spec-first）。Go サーバ型は oapi-codegen、
フロント TS 型は openapi-typescript で同じ yaml から生成する（両側手書き禁止）。

## 構成
- `backend/` Go 1.22 + Gin + oapi-codegen（strict-server）+ modernc.org/sqlite（pure Go）
- `web/` Vite + React 18 + TypeScript + MapLibre GL JS

## 実行手順

```bash
nix develop

# backend: 型生成 → ビルド
cd backend && go generate ./... && go build ./...

# API をコンテナ起動（localhost:8000）
cd .. && docker compose up -d
curl http://localhost:8000/api/pins        # pins / prefecture_count / total を含む JSON

# frontend: 型生成 → 開発サーバ（localhost:5173）
cd web
bunx openapi-typescript ../backend/openapi.yaml -o src/types.gen.ts
bun install && bun run dev
```

## ヒーロー指標
`prefecture_count` =「何都道府県に散らばっているか」。人数より広がりを主役にする。
