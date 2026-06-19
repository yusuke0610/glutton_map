# 如月愛マップ — 開発タスク
# 各ターゲットは `nix develop` のシェル内で実行する（go / bun は nix が供給）。
# まとめて nix 経由で叩きたいときは: nix develop --command make <target>

.DEFAULT_GOAL := help
.PHONY: help gen gen-backend gen-web build build-backend build-web \
        test test-backend test-web lint run dev up down clean

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

## lint: フロントの eslint
lint:
	cd web && bun install && bun run lint

## run: API をローカル起動（:8000、PORT env で変更可）
run:
	cd backend && LIBSQL_URL=file:./data/pins.db go run ./cmd/server

## dev: フロント開発サーバ（localhost:5173）
dev:
	cd web && bun install && bun run dev

## up: docker compose で API 起動（localhost:8000）
up:
	docker compose up -d --build

## down: docker compose 停止
down:
	docker compose down

## clean: ローカル DB とビルド成果物を削除
clean:
	rm -f backend/data/pins.db backend/server
	rm -rf web/dist
