# docker-compose-deploy.yaml 部署说明

这份 compose 用于服务器生产部署，只包含：

- `mysql`：MySQL 8.4
- `server`：后端镜像 `crpi-4otucz63tm2q5dhq.cn-beijing.personal.cr.aliyuncs.com/project-shiyu/agi-platform-server:latest`

用户端和管理员后台建议部署到 Cloudflare Pages，后端通过 `api.newmovieai.com` 反向代理到服务器本地 `127.0.0.1:18082`。

## 一、首次部署

在服务器安装 Docker 和 Docker Compose 后，拉取项目代码：

```bash
git clone <your-repo-url> agi-platform
cd agi-platform
```

复制环境变量模板：

```bash
cp deploy/backend.env.example deploy/backend.env
```

编辑 `deploy/backend.env`：

```bash
vim deploy/backend.env
```

至少修改这些值：

```text
MYSQL_ROOT_PASSWORD=数据库 root 密码
JWT_SECRET=很长的随机字符串
COS_SECRET_ID=腾讯云 SecretId
COS_SECRET_KEY=腾讯云 SecretKey
COS_BUCKET=腾讯云 COS 桶名
COS_REGION=腾讯云 COS 地域
COS_PUBLIC_BASE_URL=腾讯云 COS 公网访问地址
```

登录阿里云镜像仓库：

```bash
docker login crpi-4otucz63tm2q5dhq.cn-beijing.personal.cr.aliyuncs.com
```

启动：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d
```

检查：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml ps
curl http://127.0.0.1:18082/health
```

## 二、只重启后端

只重启后端容器，不拉新镜像：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml restart server
```

适合只改了环境变量、反代配置，或者想快速重启后端进程的场景。

如果改了 `deploy/backend.env`，建议用：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d server
```

## 三、拉取最新镜像并重启后端

每次 GitHub Actions 构建完成后，服务器执行：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml pull server
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d server
```

这会拉取最新的：

```text
crpi-4otucz63tm2q5dhq.cn-beijing.personal.cr.aliyuncs.com/project-shiyu/agi-platform-server:latest
```

`server` 服务设置了 `pull_policy: always`，执行 `up -d server` 时也会尽量检查并拉取最新镜像。为了结果更明确，推荐先 `pull server` 再 `up -d server`。

## 四、整体重启

重启后端和 MySQL：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml restart
```

如果想重新创建容器：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d
```

注意：不要随便执行 `down -v`，它会删除数据库 volume，导致本地 MySQL 数据丢失。

## 五、停止服务

停止容器但保留数据：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml down
```

再次启动：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d
```

## 六、查看日志

查看后端日志：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml logs -f server
```

查看 MySQL 日志：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml logs -f mysql
```

查看最近 200 行后端日志：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml logs --tail=200 server
```

## 七、查看状态

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml ps
```

后端健康检查：

```bash
curl http://127.0.0.1:18082/health
```

如果服务器外部通过 nginx 暴露了 `api.newmovieai.com`：

```bash
curl https://api.newmovieai.com/health
```

## 八、数据库数据

MySQL 数据保存在 Docker volume：

```text
mysql-data
```

后端本地兜底上传目录保存在：

```text
server-uploads
```

常规更新后端镜像不会影响数据库数据。

危险命令：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml down -v
```

这个命令会删除 `mysql-data`，生产环境不要直接执行。

## 九、反向代理

示例 nginx 配置见：

```text
deploy/nginx-api.conf.example
```

推荐外部访问：

```text
https://api.newmovieai.com
```

nginx/Caddy 反代到：

```text
http://127.0.0.1:18082
```

前端 Cloudflare Pages 中配置：

```text
VITE_API_BASE_URL=https://api.newmovieai.com
```
