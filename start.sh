#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🚀 AGI Platform 一键启动脚本"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ 错误: 未安装 Docker"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ 错误: 未安装 Docker Compose"
    echo "请先安装 Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# 复制环境变量文件
if [ ! -f .env ]; then
    echo "📝 创建环境变量文件..."
    cp .env.example .env
    echo "✅ 已创建 .env 文件，可以根据需要修改配置"
else
    echo "✅ .env 文件已存在"
fi

echo ""
echo "🔨 构建并启动所有服务..."
echo ""

# 使用 docker compose 或 docker-compose
if docker compose version &> /dev/null; then
    docker compose up -d --build
else
    docker-compose up -d --build
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  ✅ 所有服务启动成功！"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📊 服务访问地址："
    echo "  • 后端 API:    http://localhost:8080"
    echo "  • 管理后台:    http://localhost:3001"
    echo "  • MinIO 控制台: http://localhost:9001"
    echo ""
    echo "🔑 默认账号："
    echo "  • 管理员: admin / admin123"
    echo "  • MinIO: minioadmin / minioadmin123"
    echo ""
    echo "📖 查看日志："
    echo "  docker-compose logs -f [服务名]"
    echo ""
    echo "🛑 停止服务："
    echo "  docker-compose down"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
else
    echo ""
    echo "❌ 启动失败，请查看错误信息"
    exit 1
fi
