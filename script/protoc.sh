#!/usr/bin/env bash

set -eu

APP_OUTPUT_PATH=./gen/pb

# Check existance
if [ ! -d $APP_OUTPUT_PATH ]; then
  # 存在しない場合は作成
  mkdir -p $APP_OUTPUT_PATH
fi

protoc \
  --go_out=$APP_OUTPUT_PATH \
  --go_opt=paths=source_relative \
  --go-grpc_out=$APP_OUTPUT_PATH \
  --go-grpc_opt=paths=source_relative \
  -I./proto \
  ./proto/*.proto