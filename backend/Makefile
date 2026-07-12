.PHONY: help dev build run test clean docker migrate seed

# 默认目标
.DEFAULT_GOAL := help

# 项目名称
PROJECT_NAME := agi-platform-backend
API_BINARY := bin/api
WORKER_BINARY := bin/worker

# 帮助信息
help:
	@echo "AGI Platform Backend - Makefile 命令"
	@echo ""
	@echo "使用方法: make [命令]"
	@echo ""
	@echo "命令列表:"
	@echo "  dev          - 开发模式运行（热重载）"
	@echo "  build        - 构建项目"
	@echo "  run-api      - 运行 API 服务"
	@echo "  run-worker   - 运行 Worker 服务"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  docker       - 构建 Docker 镜像"
	@echo "  docker-up    - 启动 Docker 容器"
	@echo "  docker-down  - 停止 Docker 容器"
	@echo "  migrate      - 运行数据库迁移"
	@echo "  seed         - 导入种子数据"
	@echo "  deps         - 安装依赖"
	@echo "  lint         - 运行代码检查"
	@echo ""

# 安装依赖
deps:
	@echo "📦 安装依赖..."
	go mod download
	go mod tidy

# 开发模式（需要安装 air）
dev:
	@if ! command -v air > /dev/null; then \
		echo "📦 安装 air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@echo "🚀 启动开发服务器..."
	air

# 构建项目
build:
	@echo "🔨 构建项目..."
	@mkdir -p bin
	go build -o $(API_BINARY) cmd/api/main.go
	go build -o $(WORKER_BINARY) cmd/worker/main.go
	@echo "✅ 构建完成: $(API_BINARY), $(WORKER_BINARY)"

# 运行 API 服务
run-api: build
	@echo "🚀 启动 API 服务..."
	$(API_BINARY)

# 运行 Worker 服务
run-worker: build
	@echo "🚀 启动 Worker 服务..."
	$(WORKER_BINARY)

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v -cover ./...

# 运行测试（带覆盖率）
test-coverage:
	@echo "🧪 运行测试并生成覆盖率报告..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告: coverage.html"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html
	@echo "✅ 清理完成"

# 代码检查
lint:
	@if ! command -v golangci-lint > /dev/null; then \
		echo "📦 安装 golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "🔍 运行代码检查..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	go fmt ./...
	goimports -w .

# 数据库迁移
migrate:
	@echo "🗄️  运行数据库迁移..."
	@if [ -f scripts/migrations/run.sh ]; then \
		bash scripts/migrations/run.sh; \
	else \
		echo "❌ 迁移脚本不存在"; \
	fi

# 导入种子数据
seed:
	@echo "🌱 导入种子数据..."
	@if [ -f scripts/seeds/seed.sql ]; then \
		mysql -u$(DB_USER) -p$(DB_PASSWORD) -h$(DB_HOST) $(DB_NAME) < scripts/seeds/seed.sql; \
		echo "✅ 种子数据导入完成"; \
	else \
		echo "❌ 种子数据文件不存在"; \
	fi

# Docker 构建
docker:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t $(PROJECT_NAME):latest .
	@echo "✅ Docker 镜像构建完成"

# Docker Compose 启动
docker-up:
	@echo "🐳 启动 Docker 容器..."
	docker-compose up -d
	@echo "✅ 容器已启动"

# Docker Compose 停止
docker-down:
	@echo "🐳 停止 Docker 容器..."
	docker-compose down
	@echo "✅ 容器已停止"

# Docker Compose 日志
docker-logs:
	docker-compose logs -f

# 生成 API 文档（Swagger）
docs:
	@if ! command -v swag > /dev/null; then \
		echo "📦 安装 swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@echo "📚 生成 API 文档..."
	swag init -g cmd/api/main.go -o docs/swagger
	@echo "✅ API 文档生成完成"

# 查看项目信息
info:
	@echo "📊 项目信息"
	@echo "项目名称: $(PROJECT_NAME)"
	@echo "Go 版本: $$(go version)"
	@echo "模块路径: $$(go list -m)"
	@echo ""
	@echo "📦 依赖数量: $$(go list -m all | wc -l)"
	@echo "📝 代码行数: $$(find . -name '*.go' | xargs wc -l | tail -1)"
