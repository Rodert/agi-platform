# AGI Platform 开发与部署

本项目使用两套 Docker Compose 配置：`docker-compose.local.yml` 用于本地源码调试，`docker-compose.yml` 用于服务器部署已发布镜像。

## 本地调试

可选地在项目根目录创建 `.env`，配置本地数据库、端口和超级管理员：

```env
LOCAL_HTTP_PORT=3100
LOCAL_SUPER_ADMIN_USERNAME=admin
LOCAL_SUPER_ADMIN_PASSWORD=local-admin-password
LOCAL_SUPER_ADMIN_NAME=本地超级管理员
```

启动：

```bash
docker compose -f docker-compose.local.yml up
```

后台运行：

```bash
docker compose -f docker-compose.local.yml up -d
```

访问地址：

| 服务 | 地址 |
| --- | --- |
| 用户端 | `http://localhost:3100/` |
| 管理后台 | `http://localhost:3100/admin/` |
| 健康检查 | `http://localhost:3100/health` |

停止或重启服务：

```bash
docker compose -f docker-compose.local.yml down
docker compose -f docker-compose.local.yml restart api worker
```

## 服务器部署

在服务器部署目录准备 `docker-compose.yml` 和 `.env`：

```bash
cp .env.example .env
```

至少配置以下值。请使用强密码，不要提交 `.env`：

```env
MYSQL_ROOT_PASSWORD=<随机强密码>
MYSQL_PASSWORD=<随机强密码>
JWT_SECRET=<至少32位随机字符串>
SUPER_ADMIN_USERNAME=<管理员登录名>
SUPER_ADMIN_PASSWORD=<至少8位强密码>
SUPER_ADMIN_NAME=超级管理员
HTTP_PORT=3012
APP_IMAGE=ghcr.io/rodert/agi-platform:0.1.8
```

启动发布镜像：

```bash
docker login ghcr.io
docker compose pull
docker compose up -d
docker compose ps
```

访问地址：

| 服务 | 地址 |
| --- | --- |
| 用户端 | `http://<服务器IP>:3012/` |
| 管理后台 | `http://<服务器IP>:3012/admin/` |

首次启动时，应用根据 `SUPER_ADMIN_*` 创建超级管理员。若用户名已经存在，启动会保留原账号的密码和权限，不会因修改 `.env` 而重置。

更新镜像版本：

```bash
docker compose pull app
docker compose up -d --no-deps app
```

完整的镜像发布、版本检测和后台更新说明见 [DOCKER_START.md](./DOCKER_START.md)。

## 后台一键更新

服务器默认允许超级管理员在后台拉取镜像、重启或回滚。`.env.example` 已包含以下配置，部署前必须替换更新令牌：

```env
UPDATE_ENABLED=true
UPDATE_AGENT_TOKEN=<替换为至少32位随机字符串>
COMPOSE_PROFILES=update
```

随后仍使用正常启动命令：

```bash
docker compose up -d
```

`COMPOSE_PROFILES=update` 会自动启动更新代理，无需在命令行重复写 `--profile update`。更新代理可访问 Docker Socket，只应在受信任的单机服务器启用。
