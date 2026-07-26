# AGI Platform - AI 生成平台

一个完整的 AI 生成平台，包含后端服务、管理后台和用户端。

## 📁 项目结构

```
agi-platform/
├── backend/           # 后端服务（Go）
│   ├── cmd/          # 应用入口
│   ├── internal/     # 业务代码
│   ├── pkg/          # 公共库
│   ├── configs/      # 配置文件
│   ├── scripts/      # SQL脚本
│   └── bin/          # 编译产物
├── admin/            # 管理后台（Vue 3）
│   └── src/          # 前端源码
├── frontend/         # 用户端（React + Vite）
├── docker-compose.yml # Docker配置
└── start.sh          # 一键启动脚本
```

## 🚀 快速启动

### 方式 1：Docker 部署

```bash
# 启动所有服务（单应用镜像 + MySQL + Redis）
./start.sh

# 或使用 Docker Compose
docker compose up -d --build
```

部署版统一入口默认使用 `3012` 端口：用户端为 `http://localhost:3012/`，管理后台为 `http://localhost:3012/admin/`。

### 方式 2：Docker 本地调试

本地调试环境会以源码启动 API、Worker、用户端和管理端，默认使用独立端口，不影响部署环境：

```bash
docker compose -f docker-compose.local.yml up
```

- 用户端: http://localhost:3100/
- 管理后台: http://localhost:3100/admin/
- API 健康检查: http://localhost:3100/health

完整的本地调试和部署说明见 [DOCKER_START.md](./DOCKER_START.md)。

### 方式 3：手动启动

#### 启动后端
```bash
# 1. 在项目根目录启动基础服务
docker compose up -d mysql redis

# 2. 初始化数据库
mysql -u root -p agi_platform < backend/scripts/migrations/001_create_tables.sql
mysql -u root -p agi_platform < backend/scripts/seeds/seed.sql

# 3. 启动服务
cd backend
go run ./cmd/api        # API 服务
go run ./cmd/worker     # Worker 服务（另开终端）
```

#### 启动管理后台
```bash
cd admin
pnpm install
pnpm dev
```

## 📊 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 用户端 | http://localhost:3012/ | 用户界面 |
| 管理后台 | http://localhost:3012/admin/ | 管理员界面 |
| API 健康检查 | http://localhost:3012/health | 服务状态 |

## 🔑 超级管理员

首次启动时根据 `.env` 中的 `SUPER_ADMIN_USERNAME`、`SUPER_ADMIN_PASSWORD` 和 `SUPER_ADMIN_NAME` 创建超级管理员；账号存在后不会被环境变量重置。

## 🎯 核心特性

- ✅ 完整的用户认证系统
- ✅ AI 创作任务管理
- ✅ Worker 异步处理
- ✅ 作品社区（点赞/收藏）
- ✅ 灵感值、兑换码与充值套餐
- ✅ 管理后台（用户、权限、审核、配置与日志）
- ✅ Docker 一键部署

## 📖 详细文档

- [Docker 启动指南](./DOCKER_START.md) - 详细的 Docker 使用说明
- [开发与部署指南](./DEVELOP.md) - 本地调试、服务器部署与超级管理员配置
- [版本发布指南](./VERSION.md) - 版本号、GitHub Release 与镜像发布规则
- [后端文档](./backend/README.md) - 后端 API 详细文档
- [管理后台文档](./admin/README.md) - 管理后台前端文档
- [历史文档归档](./docs/archive/README.md) - 既往迁移与修复记录，仅供追溯

## 🛠️ 技术栈

### 后端
- Go 1.21
- Gin Framework
- MySQL 8.0
- Redis 7.0
- Docker Compose

### 管理后台
- Vue 3
- Vite
- Element Plus
- Pinia

## 📝 开发指南

### 后端开发
```bash
cd backend

# 安装依赖
go mod download

# 运行测试
go test ./...

# 编译
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go
```

### 前端开发
```bash
cd admin

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建生产版本
pnpm build
```

## 🐛 故障排查

### 端口被占用
```bash
# 修改 .env 文件中的部署端口配置
HTTP_PORT=3012
```

### 数据库连接失败
```bash
# 检查 MySQL 是否启动
docker compose ps mysql

# 查看日志
docker compose logs mysql
```

### 查看服务日志
```bash
# 查看所有日志
docker compose logs -f

# 查看特定服务
docker compose logs -f app
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题，请提交 Issue。
