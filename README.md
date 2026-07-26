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
├── frontend/         # 用户端（待开发）
├── docker-compose.yml # Docker配置
└── start.sh          # 一键启动脚本
```

## 🚀 快速启动

### 方式 1：Docker 部署

```bash
# 启动所有服务（单应用镜像 + MySQL + Redis）
./start.sh

# 或使用 docker-compose
docker-compose up -d --build
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
cd backend

# 1. 启动基础服务
docker-compose up -d mysql redis

# 2. 初始化数据库
mysql -u root -p agi_platform < scripts/migrations/001_create_tables.sql
mysql -u root -p agi_platform < scripts/seeds/seed.sql

# 3. 启动服务
./bin/api        # API服务
./bin/worker     # Worker服务
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
| 用户端 | http://localhost/ | 用户界面 |
| 管理后台 | http://localhost/admin/ | 管理员界面 |
| API 健康检查 | http://localhost/health | 服务状态 |

## 🔑 默认账号

### 管理员
- 用户名: `admin`
- 密码: `admin123`

## ✅ 已实现功能

### 后端 API（21个接口）

#### 用户端（17个）
- **认证模块**: 注册、登录、验证码
- **用户模块**: 获取/更新资料
- **创作模块**: 创建图片/视频任务、查询模型、任务列表
- **作品模块**: 发布作品、浏览、点赞、收藏

#### 管理后台（4个）
- **管理员登录**
- **数据统计**: 用户数、任务数、作品数
- **作品审核**: 待审核列表、审核操作

### 管理后台前端（3个页面）
- **登录页面**: 管理员认证
- **数据统计**: 可视化看板
- **作品审核**: 审核列表和操作

## 🎯 核心特性

- ✅ 完整的用户认证系统
- ✅ AI 创作任务管理
- ✅ Worker 异步处理
- ✅ 作品社区（点赞/收藏）
- ✅ 管理后台（数据统计/作品审核）
- ✅ Docker 一键部署

## 📖 详细文档

- [Docker 启动指南](./DOCKER_START.md) - 详细的 Docker 使用说明
- [后端文档](./backend/README.md) - 后端 API 详细文档
- [管理后台文档](./admin/README.md) - 管理后台前端文档

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
# 修改 .env 文件中的端口配置
HTTP_PORT=3012
```

### 数据库连接失败
```bash
# 检查 MySQL 是否启动
docker-compose ps mysql

# 查看日志
docker-compose logs mysql
```

### 查看服务日志
```bash
# 查看所有日志
docker-compose logs -f

# 查看特定服务
docker-compose logs -f backend
docker-compose logs -f worker
```

## 📄 License

MIT

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题，请提交 Issue。
