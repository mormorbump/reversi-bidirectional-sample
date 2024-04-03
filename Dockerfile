# AWS CodeBuild の Rate Limit に引っかからないようにするために ECR Public Gallery から取得する
FROM public.ecr.aws/docker/library/golang:1.22.1-alpine AS builder

ENV ROOT=/app
WORKDIR ${ROOT}

RUN apk upgrade && apk add git
COPY server server
COPY pkg pkg
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download

# 冗長性max && 重大度をinfoにしてできるだけログを出すようにする
ENV GRPC_GO_LOG_VERBOSITY_LEVEL=99
ENV GRPC_GO_LOG_SEVERITY_LEVEL=info

RUN go build server/main.go

# 開発ステージ (live-reload用)
FROM golang:1.22.1-alpine AS dev-live-reload

RUN go install github.com/cosmtrek/air@v1.49.0

# compose.yamlで`.:/app`のマウント設定が存在することを前提とする
WORKDIR /app
CMD ["air", "-c", ".air.toml"]
