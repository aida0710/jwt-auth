# JWT Auth API

Go Echo を使用したクリーンアーキテクチャベースの REST API プロジェクト

## 📚 アーキテクチャ

このプロジェクトはクリーンアーキテクチャの原則に従って設計されています：

```
jwt-auth/
├── cmd/
│   └── main.go          # アプリケーションエントリーポイント
├── internal/
│   ├── domain/          # ビジネスロジック層
│   │   ├── entity/      # ドメインモデル
│   │   └── repository/  # リポジトリインターフェース
│   ├── usecase/         # ユースケース層
│   │   ├── account/
│   │   └── project/
│   ├── infrastructure/  # インフラストラクチャ層
│   │   └── database/    # DB接続
│   ├── repository/      # リポジトリ実装
│   └── api/            # OpenAPI自動生成コード + ハンドラー実装
│       ├── types.gen.go    # 自動生成: 型定義
│       ├── server.gen.go   # 自動生成: サーバーインターフェース
│       ├── spec.gen.go     # 自動生成: 仕様
│       └── server_impl.go  # 手動実装: ハンドラー
├── api/                 # OpenAPI仕様
│   └── openapi.yaml    # OpenAPI 3.0.3定義
├── ddl/                # データベーススキーマ
│   └── schema.sql      # MySQL DDL
├── docker-compose.yml  # Docker Compose設定
├── Dockerfile          # マルチステージビルド
└── .air.toml          # ホットリロード設定
```

## 🚀 クイックスタート（Docker Compose）

### 1. 環境変数を設定
```bash
cp .env.example .env
```

### 2. Docker Composeで起動
```bash
docker-compose up -d
```

これだけで以下が自動的に実行されます：
- MySQL 8.0の起動とスキーマ作成
- OpenAPIコードの自動生成
- アプリケーションの起動（ホットリロード対応）

### 3. 動作確認
```bash
# ヘルスチェック
curl http://localhost:8080/api/v1/health

# アカウント作成
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"John Doe"}'
```

### 4. 停止
```bash
docker-compose down

# データも削除する場合
docker-compose down -v
```

## 🐳 Docker環境の詳細

### 開発環境（デフォルト）
- ホットリロード対応（Air使用）
- ソースコードをボリュームマウント
- 変更を即座に反映

### 本番環境ビルド
```bash
# 本番用イメージをビルド
docker build --target production -t jwt-auth:prod .

# 実行
docker run -p 8080:8080 --env-file .env jwt-auth:prod
```

## 📖 API エンドポイント

### Accounts

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/accounts` | アカウント作成 |
| GET | `/api/v1/accounts` | アカウント一覧取得 |
| GET | `/api/v1/accounts/{id}` | アカウント詳細取得 |
| PUT | `/api/v1/accounts/{id}` | アカウント更新 |
| DELETE | `/api/v1/accounts/{id}` | アカウント削除 |

### Projects

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/accounts/{account_id}/projects` | プロジェクト作成 |
| GET | `/api/v1/accounts/{account_id}/projects` | プロジェクト一覧取得 |
| GET | `/api/v1/accounts/{account_id}/projects/{project_id}` | プロジェクト詳細取得 |
| PUT | `/api/v1/accounts/{account_id}/projects/{project_id}` | プロジェクト更新 |
| DELETE | `/api/v1/accounts/{account_id}/projects/{project_id}` | プロジェクト削除 |

## 🔧 開発

### ローカルでの開発（Docker不使用）

1. **MySQLを起動**
```bash
mysql -u root -p
CREATE DATABASE jwt_auth;
USE jwt_auth;
source ddl/schema.sql;
```

2. **OpenAPIコード生成**
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1

oapi-codegen -generate types -package api -o internal/api/types.gen.go api/openapi.yaml
oapi-codegen -generate echo-server -package api -o internal/api/server.gen.go api/openapi.yaml
oapi-codegen -generate spec -package api -o internal/api/spec.gen.go api/openapi.yaml
```

3. **アプリケーション起動**
```bash
go mod download
go run cmd/main.go
```

### ホットリロード開発
```bash
# Air をインストール
go install github.com/air-verse/air@latest

# 起動
air -c .air.toml
```

### テスト実行
```bash
go test -v ./...
```

## 📝 環境変数

| Variable | Description | Default |
|----------|-------------|---------|
| BACKEND_PORT | サーバーポート | 8080 |
| DB_HOST | データベースホスト | db (Docker) / localhost (ローカル) |
| DB_PORT | データベースポート | 3306 |
| DB_USER | データベースユーザー | root |
| DB_PASSWORD | データベースパスワード | password |
| DB_NAME | データベース名 | jwt_auth |

## 🗂️ データモデル

### Accounts
- `id` (UUID): 主キー
- `email` (String): メールアドレス（ユニーク）
- `name` (String): 名前
- `created_at` (Timestamp): 作成日時
- `updated_at` (Timestamp): 更新日時

### Projects
- `id` (UUID): 主キー
- `account_id` (UUID): アカウントID（外部キー）
- `name` (String): プロジェクト名
- `description` (Text): 説明
- `status` (Enum): ステータス（active/inactive/archived）
- `created_at` (Timestamp): 作成日時
- `updated_at` (Timestamp): 更新日時

### リレーション
- 1つのAccountが複数のProjectを持つ（1対多）
- Account削除時、関連するProjectも削除（CASCADE DELETE）

## 🏗️ プロジェクト構成

### マルチステージDockerビルド
1. **codegen**: OpenAPIコード生成
2. **dev**: 開発環境（ホットリロード）
3. **builder**: 本番ビルド
4. **production**: 最小実行環境

### 主要ライブラリ
- [Echo](https://echo.labstack.com/) - Web フレームワーク
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - OpenAPIコード生成
- [sqlx](https://github.com/jmoiron/sqlx) - SQL ライブラリ
- [Air](https://github.com/air-verse/air) - ホットリロード
- [google/uuid](https://github.com/google/uuid) - UUID 生成
- [godotenv](https://github.com/joho/godotenv) - 環境変数管理

## 📄 ライセンス

MIT License
