# AGI Platform Backend

潮汐 AI 平台后端服务 - Go + Gin + GORM

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **对象存储**: MinIO / 阿里云 OSS / AWS S3
- **消息队列**: Redis Stream

## 功能模块

- ✅ 用户系统 - 注册、登录、个人资料
- ✅ 创作系统 - 图片/视频生成
- ✅ 任务系统 - 异步任务处理
- ✅ 作品系统 - 作品展示、点赞、收藏
- ✅ 积分系统 - 灵感值充值、消费、签到
- ✅ 支付系统 - 支持易支付、支付宝、微信支付
- ✅ 会员系统 - 会员等级、权益管理
- ✅ 邀请系统 - 邀请奖励
- ✅ 通知系统 - 站内消息、WebSocket 推送
- ✅ 管理后台 - 用户管理、作品审核、模型配置

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库、Redis 等
```

### 3. 数据库迁移

```bash
# 手动执行 SQL 迁移脚本
mysql -u root -p agi_platform < scripts/migrations/001_create_tables.sql
```

### 4. 运行服务

```bash
# 开发模式（热重载）
make dev

# 或分别启动
make run-api      # API 服务
make run-worker   # Worker 服务
```

## 目录结构

```
backend/
├── cmd/                  # 应用入口
│   ├── api/             # API 服务
│   └── worker/          # Worker 服务
├── internal/            # 私有代码
│   ├── model/          # 数据模型
│   ├── handler/        # HTTP 处理器
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据访问
│   ├── middleware/     # 中间件
│   └── worker/         # Worker 处理器
├── pkg/                # 公共库
│   ├── config/        # 配置管理
│   ├── database/      # 数据库连接
│   ├── logger/        # 日志
│   ├── jwt/           # JWT
│   └── utils/         # 工具函数
├── configs/           # 配置文件
└── scripts/           # 脚本
```

## 开发命令

```bash
make help          # 查看所有命令
make deps          # 安装依赖
make dev           # 开发模式
make build         # 构建项目
make test          # 运行测试
make lint          # 代码检查
make docker        # 构建 Docker 镜像
```

## API 文档

启动服务后访问：
- Swagger UI: http://localhost:8080/swagger/index.html
- API 文档: `/docs/api/`

## 环境变量

主要配置项：

```env
# 数据库
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=agi_platform

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key

# 对象存储
STORAGE_TYPE=minio
STORAGE_ENDPOINT=localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
```

## 部署

### Docker 部署

```bash
# 构建镜像
make docker

# 启动服务
make docker-up
```

### 生产环境

```bash
# 构建
make build

# 运行
./bin/api --config=configs/config.prod.yaml
./bin/worker --config=configs/config.prod.yaml
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
