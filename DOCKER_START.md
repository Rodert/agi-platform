# 🚀 AGI Platform - 一键启动指南

## 快速启动（推荐）

### 方式 1：使用启动脚本（最简单）

```bash
# 一键启动所有服务
./start.sh
```

### 方式 2：使用 Docker Compose

```bash
# 1. 复制环境变量配置
cp .env.example .env

# 2. 启动所有服务
docker-compose up -d --build

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f
```

---

## 📊 服务访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| **后端 API** | http://localhost:8080 | API 接口 |
| **管理后台** | http://localhost:3001 | 管理员后台 |
| **MinIO 控制台** | http://localhost:9001 | 对象存储管理 |
| **MySQL** | localhost:3306 | 数据库 |
| **Redis** | localhost:6379 | 缓存 |

---

## 🔑 默认账号

### 管理员
- 用户名: `admin`
- 密码: `admin123`

### MinIO
- Access Key: `minioadmin`
- Secret Key: `minioadmin123`

### MySQL
- Root 密码: `root123456`
- 用户: `agi`
- 密码: `agi123456`

---

## 📦 包含的服务

```
┌─────────────────────────────────────────┐
│  AGI Platform 服务架构                   │
├─────────────────────────────────────────┤
│                                          │
│  ┌──────────┐  ┌──────────┐            │
│  │  管理后台  │  │  后端API  │            │
│  │  :3001   │──│  :8080   │            │
│  └──────────┘  └──────────┘            │
│                     │                   │
│       ┌─────────────┼─────────────┐    │
│       │             │             │    │
│  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐│
│  │  MySQL  │  │  Redis  │  │  MinIO  ││
│  │  :3306  │  │  :6379  │  │  :9000  ││
│  └─────────┘  └─────────┘  └─────────┘│
│                                          │
│  ┌──────────┐                           │
│  │  Worker  │  (后台任务处理)            │
│  └──────────┘                           │
│                                          │
└─────────────────────────────────────────┘
```

---

## 🛠️ 常用命令

### 启动服务
```bash
docker-compose up -d
```

### 停止服务
```bash
docker-compose down
```

### 重启服务
```bash
docker-compose restart
```

### 查看服务状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f worker
docker-compose logs -f admin
```

### 进入容器
```bash
# 进入后端容器
docker-compose exec backend sh

# 进入数据库容器
docker-compose exec mysql bash
```

### 清理数据（危险操作）
```bash
# 停止并删除所有容器和数据卷
docker-compose down -v
```

---

## 🔧 自定义配置

### 修改端口

编辑 `.env` 文件：

```env
API_PORT=8080        # 后端 API 端口
ADMIN_PORT=3001      # 管理后台端口
MYSQL_PORT=3306      # MySQL 端口
REDIS_PORT=6379      # Redis 端口
```

### 修改密码

编辑 `.env` 文件：

```env
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_db_password
MINIO_ROOT_USER=your_minio_user
MINIO_ROOT_PASSWORD=your_minio_password
```

---

## 📖 初始化数据

数据库会在首次启动时自动初始化。

如需手动初始化：

```bash
# 进入 MySQL 容器
docker-compose exec mysql bash

# 执行 SQL
mysql -u root -p agi_platform < /docker-entrypoint-initdb.d/001_create_tables.sql
```

---

## 🐛 故障排查

### 服务启动失败

```bash
# 查看服务日志
docker-compose logs [服务名]

# 检查端口占用
lsof -i :[端口号]
```

### 数据库连接失败

```bash
# 检查 MySQL 是否启动
docker-compose ps mysql

# 查看 MySQL 日志
docker-compose logs mysql
```

### 清理并重新启动

```bash
# 停止所有服务
docker-compose down

# 清理旧容器和镜像
docker system prune -a

# 重新构建并启动
docker-compose up -d --build
```

---

## ✅ 验证服务

### 1. 检查健康状态
```bash
# API 健康检查
curl http://localhost:8080/health

# 预期输出
{"status":"ok","message":"服务运行正常"}
```

### 2. 访问管理后台
浏览器打开：http://localhost:3001

使用默认账号登录：`admin` / `admin123`

### 3. 查看统计数据
登录后会看到数据统计看板

---

## 📝 注意事项

1. **首次启动较慢** - 需要下载镜像和构建服务
2. **数据持久化** - 数据存储在 Docker 卷中，停止服务不会丢失数据
3. **端口冲突** - 确保所需端口未被占用
4. **资源占用** - 建议至少 4GB 内存

---

## 🎉 启动成功后

访问 http://localhost:3001 即可使用管理后台！

- 查看数据统计
- 审核用户作品
- 管理系统配置

---

**祝你使用愉快！** 💪
