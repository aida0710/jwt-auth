.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## OpenAPI仕様からコードを生成（複数ファイル）
	@echo "Generating OpenAPI code..."
	@oapi-codegen -generate types -package api -o internal/api/types.gen.go api/openapi.yaml
	@oapi-codegen -generate echo-server -package api -o internal/api/server.gen.go api/openapi.yaml
	@oapi-codegen -generate spec -package api -o internal/api/spec.gen.go api/openapi.yaml
	@echo "OpenAPI code generation completed!"

.PHONY: install-tools
install-tools: ## 必要なツールをインストール
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/air-verse/air@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: run
run: ## アプリケーションを実行
	go run cmd/main.go

.PHONY: dev
dev: ## 開発モード（ホットリロード）で実行
	air

.PHONY: build
build: ## アプリケーションをビルド
	go build -o bin/app cmd/main.go

.PHONY: test
test: ## テストを実行
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## テストカバレッジを表示
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Lintを実行
	golangci-lint run ./...

.PHONY: fmt
fmt: ## コードをフォーマット
	go fmt ./...
	gofmt -s -w .

.PHONY: clean
clean: ## ビルド成果物をクリーン
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out

.PHONY: migrate
migrate: ## データベースマイグレーションを実行
	docker exec -i db mysql -uroot -ppassword jwt_auth < ddl/schema.sql

.PHONY: docker-up
docker-up: ## Docker Composeでサービスを起動
	docker compose up -d

.PHONY: docker-down
docker-down: ## Docker Composeでサービスを停止
	docker compose down

.PHONY: docker-logs
docker-logs: ## Dockerログを表示
	docker compose logs -f

.PHONY: docker-rebuild
docker-rebuild: ## Dockerイメージを再ビルド
	docker compose build --no-cache

.PHONY: setup
setup: install-tools generate ## 初期セットアップ（ツールインストール＋コード生成）
	@echo "セットアップ完了！"

.PHONY: mock-generate
mock-generate: ## モックを生成
	mockgen -source=internal/domain/repository/account_repository.go -destination=internal/mock/account_repository_mock.go -package=mock
	mockgen -source=internal/domain/repository/project_repository.go -destination=internal/mock/project_repository_mock.go -package=mock
	mockgen -source=internal/usecase/usecases.go -destination=internal/mock/usecase_mock.go -package=mock

.PHONY: api-doc
api-doc: ## OpenAPIドキュメントをブラウザで開く
	@echo "Opening API documentation..."
	@python3 -m http.server 8090 --directory api &
	@open http://localhost:8090/openapi.yaml

.PHONY: check
check: fmt lint test ## フォーマット、Lint、テストを実行
	@echo "All checks passed!"
