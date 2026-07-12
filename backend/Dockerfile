# 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 API
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api cmd/api/main.go

# 构建 Worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/worker cmd/worker/main.go

# API 服务镜像
FROM alpine:latest AS api
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/api .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./api"]

# Worker 服务镜像
FROM alpine:latest AS worker
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/worker .
COPY --from=builder /app/configs ./configs
CMD ["./worker"]
