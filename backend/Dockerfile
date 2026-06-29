# syntax=docker/dockerfile:1.7

FROM golang:1.26.4 AS builder
WORKDIR /src

# 设置 Go 代理，解决国内服务器拉取依赖超时的问题
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

FROM ubuntu:24.04
RUN groupadd -r app && useradd -r -g app -u 65532 -s /usr/sbin/nologin app
WORKDIR /app
COPY --from=builder /out/app /app/app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
