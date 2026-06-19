# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

如月愛マップ — 日本地図（国土地理院 淡色タイル）上に API から取得したピンをヒートマップ描画する縦割り1スライス。ヒーロー指標は `prefecture_count`（人数ではなく「何都道府県に散らばっているか」）。

## 絶対原則

- **spec-first / 単一の真実**: `backend/openapi.yaml` が唯一の契約。Go サーバ型は oapi-codegen、フロント TS 型は openapi-typescript で **同じ yaml から生成する**。**両側の型を手書きしてはいけない**（`internal/api/gen.go` と `web/src/types.gen.ts` は生成物）。
- **DB の隔離**: `database/sql` とドライバ（`modernc.org/sqlite`）を import してよいのは `backend/internal/pin/repository.go` **だけ**。他層は `PinRepository` interface 越しにアクセスする。`grep -rl "database/sql" internal cmd` が1ファイルに収まることを保つ。
- **最小スコープ**: スコープ外（実装しない）= LLMモデレーション / Turnstile / 通報 / PostGIS等の高度集計 / クラスタリング / ピン個別ポップアップ / go-libsql(cgo)実装 / `weight` カラム。緯度経度はただのカラム、密度は件数で表現する。
- **TDD 遵守**: 機能追加・変更は必ず **TDD（Red→Green→Refactor）** で進める。まず失敗するテストを書いて赤を確認し、最小実装で緑にし、緑を保ったままリファクタする。**実装を先に書いてはいけない**。テストは `make test` 経由で実行する（詳細は「テスト」節）。
- **push / PR は勝手にやらない**: `git push` と PR 作成（`gh pr create` 等）は、ユーザーが明示的に指示したときだけ実行する。ローカルでの commit までは進めてよいが、リモートへ反映する操作は必ず事前に許可を取る。

## 開発コマンド

すべて `nix develop` のシェル内で実行する（`go` / `bun` は PATH になく nix が供給）。flake はファイルが git に **追跡されている**必要があるため、新規ファイルは `git add` してから nix コマンドを叩く（コミットは不要）。

```bash
nix develop

# バックエンド: 型生成 → ビルド
cd backend && go generate ./... && go build ./...

# ローカル起動（既定 :8000、PORT env で変更可）
LIBSQL_URL=file:./data/pins.db go run ./cmd/server
curl http://localhost:8000/api/pins

# フロント: 型生成 → 開発サーバ / ビルド
cd web
bunx openapi-typescript ../backend/openapi.yaml -o src/types.gen.ts
bun install && bun run dev      # localhost:5173
bun run build                   # tsc + vite（型チェック込み）

# コンテナで API 起動
docker compose up -d            # localhost:8000
```

### テスト

このプロダクトは **TDD（テスト駆動開発）** で進める。Red（失敗するテストを先に書く）→ Green（最小実装で通す）→ Refactor のサイクルを回す。テスト実行は必ず Makefile 経由で行う。

```bash
make test          # backend + web 両方
make test-backend  # go test -race -count=1 ./...
make test-web      # フロント vitest（cd web && bun run test）
make lint          # = lint-backend + lint-web
make lint-backend  # golangci-lint run ./...
make lint-web      # フロント eslint（cd web && bun run lint）
```

backend は Go 標準 `testing`（`-race`/`-count=1`）＋ **golangci-lint**（`backend/.golangci.yml`、govet/staticcheck/errcheck 等）。フロントは **vitest**（`web/src/*.test.ts`）でロジックをテストし、**eslint**（flat config: `web/eslint.config.js`）で静的検査する。これらは GitHub Actions（`.github/workflows/ci.yml`）の PR で自動実行される（backend-test / backend-lint / web-test の3ジョブ）。検証は build + curl + ブラウザ目視に加え、上記のテストで行う。

## アーキテクチャ

```
backend/
  openapi.yaml          # 契約（起点）。変更したら go generate と web の型生成を両方やり直す
  generate.go           # package tools。//go:generate をモジュールルートに置き、相対パス解決を成立させる
  tools.go              # oapi-codegen を go.mod に固定（//go:build tools）
  internal/api/         # gen.go(生成: StrictServerInterface) + handler.go(実装)
  internal/pin/         # pin.go(ドメイン) + repository.go(DBを知る唯一の場所/seam)
  internal/db/seed.go   # 重心+ゆらぎで pins 投入。DB空のとき起動時に流す
  cmd/server/main.go    # Gin + CORS(5173) + seed + NewStrictHandler でラップして登録
web/
  src/{App.tsx,api.ts,types.gen.ts}  # MapLibre heatmap。api.ts は生成型 components["schemas"][...] を参照
```

データの流れ: `main` が DSN(`LIBSQL_URL`)で `NewSQLiteRepository` → 空なら `db.Seed` → `Handler.GetApiPins` が repo から取得し `prefecture_count`(distinct)/`total` を集計 → strict-server の型付きレスポンス `GetApiPins200JSONResponse` で返す。

### コード生成の要点

- oapi-codegen は `strict-server: true`。ハンドラは型付きシグネチャ `GetApiPins(ctx, GetApiPinsRequestObject) (GetApiPinsResponseObject, error)` で実装し、生のステータスコード操作はしない。
- `//go:generate` は `backend/generate.go`（`package tools`、作業ディレクトリ = モジュールルート）に置く。`gen.go` 自身に置くと生成で上書きされ消えるため。

## 既知の差分・注意

- flake は `go`（nixpkgs unstable に `go_1_22` が無いため）。go.mod は `go 1.22` 互換のまま。
- 将来 Turso embedded（go-libsql, cgo）へ移行する際は、`PinRepository` に `NewLibsqlRepository` を足して main で差し替えるだけにする。cgo 採用時は scratch イメージが使えないため Dockerfile を distroless 等へ変更すること。
