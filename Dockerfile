# ステージ1: OpenAPIコード生成
FROM golang:1.24-alpine AS codegen

WORKDIR /app

# OAPIコードジェネレーターをインストール
ENV OAPI_CODEGEN_VERSION=v2.5.0
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}

# 必要なファイルのみコピー
COPY api/openapi.yaml ./api/
COPY go.mod go.sum ./

# OpenAPIコード生成
RUN mkdir -p internal/api && \
    oapi-codegen -generate types -package api -o internal/api/types.gen.go ./api/openapi.yaml && \
    oapi-codegen -generate echo-server -package api -o internal/api/server.gen.go ./api/openapi.yaml && \
    oapi-codegen -generate spec -package api -o internal/api/spec.gen.go ./api/openapi.yaml

# ステージ2: 開発環境
FROM golang:1.24-alpine AS dev

WORKDIR /app

# 必要なツールをインストール
RUN apk add --no-cache tzdata

ENV TZ=Asia/Tokyo
ENV GO111MODULE=on

# airをインストール（ホットリロード）
RUN go install github.com/air-verse/air@latest

# ソースコードをコピー
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CodeGenステージから生成コードをコピー
COPY --from=codegen /app/internal/api/*.gen.go ./internal/api/

ARG PORT=8080
ENV BACKEND_PORT=${PORT}
EXPOSE ${PORT}

CMD ["air", "-c", ".air.toml"]
