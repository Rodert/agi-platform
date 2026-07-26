#!/bin/bash

# AGI Platform Backend 项目验证脚本

echo "🔍 开始验证项目..."
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warning() {
    echo -e "${YELLOW}!${NC} $1"
}

# 检查 Go 版本
echo "📦 检查 Go 环境..."
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    success "Go 已安装: $GO_VERSION"
else
    error "Go 未安装，请安装 Go 1.21+"
    exit 1
fi

# 检查依赖
echo ""
echo "📦 检查依赖..."
if [ -f "go.mod" ]; then
    success "go.mod 存在"
    if [ -f "go.sum" ]; then
        success "go.sum 存在"
    else
        warning "go.sum 不存在，运行 go mod tidy"
    fi
else
    error "go.mod 不存在"
    exit 1
fi

# 检查目录结构
echo ""
echo "📁 检查目录结构..."
dirs=("cmd/api" "cmd/worker" "internal/model" "internal/handler" "internal/service" "internal/repository" "internal/middleware" "pkg/config" "pkg/database" "pkg/logger" "configs" "scripts/migrations")
for dir in "${dirs[@]}"; do
    if [ -d "$dir" ]; then
        success "$dir"
    else
        error "$dir 不存在"
    fi
done

# 检查配置文件
echo ""
echo "⚙️  检查配置文件..."
if [ -f "configs/config.yaml" ]; then
    success "configs/config.yaml 存在"
else
    error "configs/config.yaml 不存在"
fi

if [ -f ".env.example" ]; then
    success ".env.example 存在"
    if [ ! -f ".env" ]; then
        warning ".env 不存在，请复制 .env.example 为 .env"
    else
        success ".env 已配置"
    fi
else
    error ".env.example 不存在"
fi

# 检查数据库迁移脚本
echo ""
echo "🗄️  检查数据库迁移脚本..."
if [ -f "scripts/migrations/001_create_tables.sql" ]; then
    success "迁移脚本存在"
    TABLE_COUNT=$(grep -c "CREATE TABLE" scripts/migrations/001_create_tables.sql)
    success "包含 $TABLE_COUNT 张表"
else
    error "迁移脚本不存在"
fi

# 检查种子数据
if [ -f "scripts/seeds/seed.sql" ]; then
    success "种子数据存在"
else
    warning "种子数据不存在"
fi

# 检查核心文件
echo ""
echo "📄 检查核心文件..."
core_files=(
    "cmd/api/main.go"
    "cmd/worker/main.go"
    "pkg/config/config.go"
    "pkg/database/mysql.go"
    "pkg/database/redis.go"
    "pkg/logger/logger.go"
    "pkg/jwt/jwt.go"
    "pkg/errors/errors.go"
    "pkg/response/response.go"
)

for file in "${core_files[@]}"; do
    if [ -f "$file" ]; then
        success "$file"
    else
        error "$file 不存在"
    fi
done

# 统计代码
echo ""
echo "📊 代码统计..."
GO_FILES=$(find . -name "*.go" -type f | wc -l | tr -d ' ')
SQL_FILES=$(find . -name "*.sql" -type f | wc -l | tr -d ' ')
success "Go 文件: $GO_FILES 个"
success "SQL 文件: $SQL_FILES 个"

# 检查模型文件
echo ""
echo "🗂️  检查数据模型..."
model_files=(
    "internal/model/user.go"
    "internal/model/task.go"
    "internal/model/work.go"
    "internal/model/credit.go"
    "internal/model/payment.go"
    "internal/model/invitation.go"
    "internal/model/admin.go"
    "internal/model/config.go"
)

for file in "${model_files[@]}"; do
    if [ -f "$file" ]; then
        success "$(basename $file)"
    else
        error "$(basename $file) 不存在"
    fi
done

# 尝试编译
echo ""
echo "🔨 尝试编译..."
if go build -o /tmp/agi-api cmd/api/main.go 2>/dev/null; then
    success "API 编译成功"
    rm -f /tmp/agi-api
else
    error "API 编译失败"
fi

if go build -o /tmp/agi-worker cmd/worker/main.go 2>/dev/null; then
    success "Worker 编译成功"
    rm -f /tmp/agi-worker
else
    error "Worker 编译失败"
fi

# 检查 Docker 文件
echo ""
echo "🐳 检查 Docker 配置..."
if [ -f "Dockerfile" ]; then
    success "Dockerfile 存在"
else
    warning "Dockerfile 不存在"
fi

if [ -f "docker-compose.yml" ]; then
    success "docker-compose.yml 存在"
else
    warning "docker-compose.yml 不存在"
fi

# 总结
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 项目验证完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 下一步操作："
echo "1. 配置环境变量: cp .env.example .env && vi .env"
echo "2. 启动数据库: docker-compose up -d mysql redis"
echo "3. 执行迁移: mysql -u root -p agi_platform < scripts/migrations/001_create_tables.sql"
echo "4. 导入种子数据: mysql -u root -p agi_platform < scripts/seeds/seed.sql"
echo "5. 启动服务: make dev 或 go run cmd/api/main.go"
echo ""
echo "📖 更多信息请查看: README.md 和 PROJECT_SUMMARY.md"
