# Docker 使用指南

项目提供两套 Compose 配置：

| 文件 | 适用场景 | 服务形态 |
| --- | --- | --- |
| `docker-compose.yml` | 部署 | 单个发布镜像 + MySQL + Redis |
| `docker-compose.local.yml` | 本地开发 | 源码热更新的 API、Worker、用户端、管理端 + MySQL + Redis |

## 本地调试

```bash
docker compose -f docker-compose.local.yml up
```

首次启动会下载依赖，后续会复用 Docker 卷中的依赖缓存。本地调试也通过单个网关入口访问：

| 服务 | 地址 |
| --- | --- |
| 用户端 | `http://localhost:3100/` |
| 管理后台 | `http://localhost:3100/admin/` |
| API 健康检查 | `http://localhost:3100/health` |

MySQL、Redis、API、用户端和管理端只在 Docker 内部网络开放；前端、管理端通过本地网关访问 API。修改前端源码会热更新，修改 Go 源码后重启对应服务即可：

```bash
docker compose -f docker-compose.local.yml restart api worker
```

停止本地调试环境：

```bash
docker compose -f docker-compose.local.yml down
```

本地调试数据独立保存在 `agi-platform-local_*` 卷中。需要完全重置时才执行 `docker compose -f docker-compose.local.yml down -v`。

# Docker 部署指南

项目以 3 个容器运行：

| 容器 | 镜像 | 职责 |
| --- | --- | --- |
| `app` | `agi-platform` | 用户端、管理端、Go API 与 Worker |
| `mysql` | `mysql:8.0` | 数据库 |
| `redis` | `redis:7-alpine` | 缓存和任务队列 |

应用镜像中的统一入口：

| 地址 | 服务 |
| --- | --- |
| `http://<host>/` | 用户端 |
| `http://<host>/admin/` | 管理后台 |
| `http://<host>/health` | API 健康检查 |

## 首次部署

```bash
cp .env.example .env
```

编辑 `.env`，替换以下值为随机强密码：

```env
MYSQL_ROOT_PASSWORD=...
MYSQL_PASSWORD=...
JWT_SECRET=...
```

本机构建并启动：

```bash
docker compose up -d --build
docker compose ps
```

首次创建 MySQL 数据卷时，容器会依次执行 `backend/scripts/migrations/` 中的迁移和种子数据。已有数据卷不会重复初始化。

## GitHub 自动构建

工作流 [docker-image.yml](./.github/workflows/docker-image.yml) 会在任意分支收到 `git push` 后构建并推送带提交 SHA 的应用镜像；默认分支额外更新 `latest`：

```text
ghcr.io/<GitHub 用户或组织>/agi-platform:latest
ghcr.io/<GitHub 用户或组织>/agi-platform:<版本号>
ghcr.io/<GitHub 用户或组织>/agi-platform:sha-<完整提交 SHA>
```

GitHub 无法感知纯本地 `git commit`，因此必须推送：

```bash
git push origin <branch>
```

首次使用 GHCR 时，仓库的 Actions 必须具备 `packages: write` 权限；该工作流已经声明所需权限。服务器部署已发布的镜像时，在 `.env` 设置：

```env
APP_IMAGE=ghcr.io/<GitHub 用户或组织>/agi-platform:latest
```

然后执行：

```bash
docker login ghcr.io
docker compose pull app
docker compose up -d
```

## 版本发布与后台更新检测

根目录 `VERSION` 是发布版本的唯一来源，当前从 `0.1.0` 开始。修改该文件后推送到默认分支，工作流会创建对应的 GitHub Release，并同时推送同名镜像标签。管理后台左上角显示当前版本，会自动检查 GitHub Release；点击版本号可手动再次检测。

## 日常运维

```bash
docker compose logs -f app
docker compose restart app
docker compose down
```

数据库、Redis 和上传文件分别位于 `mysql_data`、`redis_data`、`uploads_data` 具名卷。`docker compose down` 不会删除数据；`docker compose down -v` 会删除所有持久化数据。
