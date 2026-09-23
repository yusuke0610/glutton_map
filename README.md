# glutton_map

高知県発祥のソウルフード**くいしんぼ如月**をこよなく愛する人あなたが住んでいる場所にピンを打ち日本全土から如月にラブコールを送るアプリです。

## セットアップ

1. `nix develop` でシェルに入る（`go` / `bun` は nix が供給する）。
2. ルートに `.env.example`、`web/` に `.env.example` があるので、それぞれ `.env` / `.env.local` としてコピーし、必要な値を埋める（詳細は「環境変数」を参照）。
3. `make help` で利用可能なターゲットの一覧を確認する。

開発タスク（型生成・ビルド・コンテナ起動・テスト・Lint 等）はすべて **`Makefile` に集約**している。
個別のコマンドは README には書かず、`Makefile` を唯一の入口とする。

## 環境変数

- **バックエンド（docker-compose 経由の本番相当の起動）**: ルートの `.env.example` を参照。`IP_HASH_SALT` は必須（未設定だと起動失敗する）。他はローカル開発向けの既定値がある。
- **バックエンド（`make run` によるローカル起動）**: `IP_HASH_SALT` は Makefile が既定値を渡すため追加設定は不要。本番相当の値で試したい場合は `IP_HASH_SALT=独自の値` を前置きする。
- **フロントエンド**: `web/.env.example` を参照。`VITE_API_BASE` は必須（未設定だとアプリ起動時にエラーになる）。`web/.env.local` としてコピーして使う。

## アーキテクチャ

Go(Gin) + SQLite のバックエンドと、React + MapLibre のフロントで構成する縦割りスライス。
契約は `backend/openapi.yaml` を唯一の真実とし、両側の型をここから生成する。
ディレクトリ構成やデータフローの詳細は `.claude/CLAUDE.md` を参照。
