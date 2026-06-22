# glutton_map — 開発タスク
# 各ターゲットは `nix develop` のシェル内で実行する（go / bun は nix が供給）。
# まとめて nix 経由で叩きたいときは: nix develop --command make <target>

.DEFAULT_GOAL := help
.PHONY: help gen gen-backend gen-web gen-kana gen-municipalities build build-backend build-web \
        test test-backend test-web test-e2e lint lint-backend lint-web run dev up down clean

## help: ターゲット一覧を表示
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

## gen: バックエンド・フロント両方の型を openapi.yaml から再生成
gen: gen-backend gen-web

## gen-backend: oapi-codegen で internal/api/gen.go を生成
gen-backend:
	cd backend && go generate ./...

## gen-web: openapi-typescript で web/src/types.gen.ts を生成
gen-web:
	cd web && bunx openapi-typescript ../backend/openapi.yaml -o src/types.gen.ts

## gen-kana: 総務省CSVから KANA_CSV(code,kana) を生成（要 KANA_SRC=全国版.csv,政令市区版.csv）
gen-kana:
	cd tools/municipalities && KANA_SRC=$(KANA_SRC) KANA_OUT=$(KANA_OUT) node prep-kana.mjs

## gen-municipalities: 行政区域(N03)から市区町村データを生成（要 N03_GEOJSON）
gen-municipalities:
	cd tools/municipalities && bun install && N03_GEOJSON=$(N03_GEOJSON) SIMPLIFY=$(SIMPLIFY) KANA_CSV=$(KANA_CSV) node generate.mjs

## build: バックエンド・フロント両方をビルド
build: build-backend build-web

## build-backend: 生成 → go build
build-backend: gen-backend
	cd backend && go build ./...

## build-web: 生成 → bun install → tsc + vite build
build-web: gen-web
	cd web && bun install && bun run build

## test: バックエンド・フロント両方のテストを実行
test: test-backend test-web

## test-backend: go test（-race / -count=1 で実行）
test-backend:
	cd backend && go test -race -count=1 ./...

## test-web: フロントのテスト（vitest）
test-web:
	cd web && bun install && bun run test

## test-e2e: E2E（Playwright）。backend+frontend を起動して縦割りを通す
test-e2e:
	cd web && bun install && bunx playwright install chromium && bun run test:e2e

## lint: バックエンド・フロント両方の静的解析
lint: lint-backend lint-web

## lint-backend: golangci-lint（govet/staticcheck/errcheck 等）
lint-backend:
	cd backend && golangci-lint run ./...

## lint-web: フロントの eslint
lint-web:
	cd web && bun install && bun run lint

## run: API をローカル起動（:8001、PORT env で変更可）
run:
	cd backend && LIBSQL_URL=file:./data/pins.db go run ./cmd/server

## dev: フロント開発サーバ（localhost:5174）
dev:
	cd web && bun install && bun run dev

## up: docker compose で API 起動（localhost:8001）
up:
	docker compose up -d --build

## down: docker compose 停止
down:
	docker compose down

## clean: ローカル DB とビルド成果物を削除
clean:
	rm -f backend/data/pins.db backend/server
	rm -rf web/dist
