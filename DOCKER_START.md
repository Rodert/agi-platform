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

部署版通过宿主机 `3012` 端口提供统一入口；容器内部仍使用 `80`，不对宿主机直接开放。

| 地址 | 服务 |
| --- | --- |
| `http://<host>:3012/` | 用户端 |
| `http://<host>:3012/admin/` | 管理后台 |
| `http://<host>:3012/health` | API 健康检查 |

## 首次部署

```bash
cp .env.example .env
```

编辑 `.env`，替换以下值为随机强密码：

```env
MYSQL_ROOT_PASSWORD=...
MYSQL_PASSWORD=...
JWT_SECRET=...
SUPER_ADMIN_USERNAME=...
SUPER_ADMIN_PASSWORD=...
HTTP_PORT=3012
```

本机构建并启动：

```bash
docker compose up -d --build
docker compose ps
```

首次创建 MySQL 数据卷时，容器会按文件名顺序执行 `backend/scripts/migrations/` 中的迁移，再导入种子数据。应用随后根据 `.env` 的 `SUPER_ADMIN_USERNAME`、`SUPER_ADMIN_PASSWORD` 和 `SUPER_ADMIN_NAME` 创建超级管理员；账号已存在时不会重置。已有数据卷不会重复初始化。

后续镜像更新会在启动应用前执行未记录的增量迁移，并将文件名写入 `schema_migrations`。更新代理会先运行迁移任务，迁移失败时不会替换正在运行的应用；手工执行 `docker compose up -d` 时，应用入口也会执行同一检查。首次升级到具备此机制的已有数据库会将 `001-022` 视为历史基线，只执行之后新增的迁移。

## GitHub 自动构建

工作流 [docker-image.yml](./.github/workflows/docker-image.yml) 会在任意分支收到 `git push` 后构建并推送带提交 SHA 的应用镜像；默认分支额外更新 `latest`：

```text
ghcr.io/<GitHub 用户或组织>/agi-platform:latest
ghcr.io/<GitHub 用户或组织>/agi-platform:<版本号>
ghcr.io/<GitHub 用户或组织>/agi-platform:sha-<完整提交 SHA>
```

每个镜像标签同时包含 `linux/amd64` 和 `linux/arm64`。Docker 会按部署机器的架构自动选择镜像，无需在标签中区分架构；不构建 Windows 或 macOS 容器镜像。

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

根目录 `VERSION` 是发布版本的唯一来源，当前从 `0.1.0` 开始。修改该文件后推送到默认分支，工作流会创建对应的 GitHub Release，并同时推送同名镜像标签。管理后台左上角显示当前版本；打开版本管理可读取后端缓存，点击“刷新版本”可手动再次检测。

版本检查由后端访问 GitHub，并以 Redis 缓存结果一小时。页面刷新不会触发 GitHub 请求；打开版本管理读取缓存，只有“刷新版本”会强制重新检查。匿名 GitHub API 的额度较低，生产环境建议在 `.env` 配置一个仅有仓库只读权限的 `GITHUB_TOKEN`。

### 后台一键更新

部署默认允许超级管理员从后台直接拉取并重启 `app`。`.env.example` 已包含更新配置，部署前必须替换 `UPDATE_AGENT_TOKEN`，并使用发布镜像而非本地构建镜像：

```env
APP_IMAGE=ghcr.io/rodert/agi-platform:latest
UPDATE_ENABLED=true
UPDATE_AGENT_TOKEN=<替换为至少 32 位随机字符串>
COMPOSE_PROFILES=update
```

`.env` 中的 `COMPOSE_PROFILES=update` 会自动启用更新代理，因此正常执行 `docker compose up -d` 即可。后台检测到新版时会显示“立即更新”。更新代理拥有 Docker Socket 权限，只应部署在受信任的单机环境，并保持令牌私密。

## 日常运维

```bash
docker compose logs -f app
docker compose restart app
docker compose down
```

数据库、Redis 和上传文件分别位于 `mysql_data`、`redis_data`、`uploads_data` 具名卷。`docker compose down` 不会删除数据；`docker compose down -v` 会删除所有持久化数据。
