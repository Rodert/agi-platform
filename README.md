# AGI Platform

AGI Platform 是一个面向 AI 图片和视频生成的 Web 平台，包含用户端、管理员后台和 Go 后端服务。

当前功能覆盖：

- 用户注册、登录、API Key 管理
- AI 生图、生视频任务提交、后台执行、任务详情与历史记录
- 参考图片、视频、音频上传到腾讯云 COS
- 上游 API Provider、API Key、图片模型、视频模型统一管理
- 用户积分余额、加减积分、积分流水
- 管理员任务列表、任务详情、请求参数查看与 JSON 复制
- 管理员数据表浏览、字段中文说明、DDL 查看、记录详情弹窗
- GitHub Actions 自动构建并推送后端 Docker 镜像

## 技术栈

- 后端：Go、Gin、GORM、MySQL
- 前端：Vue 3、Vite、TypeScript
- 存储：腾讯云 COS
- 部署：Docker Compose、GitHub Actions、阿里云 ACR

## 目录结构

```text
.
├── server                 # Go 后端
├── web
│   ├── user               # 用户端
│   ├── admin              # 管理员后台
│   └── shared             # 前端共享 API client 和类型
├── database/mysql         # 初始化 SQL
├── docs                   # 设计文档
├── config.yaml            # 本地配置示例
├── config.yaml.md         # 配置中文说明
├── compose.yaml           # 开发环境 compose
├── docker-compose-deploy.yaml # 服务器部署 compose
└── .github/workflows      # Docker 镜像构建 CI
```

## 本地启动

使用开发环境 Docker Compose：

```bash
docker compose up -d
```

服务地址：

- 用户端：http://localhost:5173
- 管理员后台：http://localhost:5174
- 后端 API：http://localhost:8080
- 健康检查：http://localhost:8080/health

查看日志：

```bash
docker compose logs -f server
docker compose logs -f user-web
docker compose logs -f admin-web
```

停止服务：

```bash
docker compose down
```

清理容器和数据卷：

```bash
docker compose down -v
```

## 默认账号

| 入口 | 地址 | 账号 | 密码 | 来源 |
| --- | --- | --- | --- | --- |
| 管理员后台 | `http://localhost:5174` | `admin` | `admin123` | `config.yaml` 的 `admin` 配置 |
| 用户端 | `http://localhost:5173` | `user@example.com` | `secret123` | `database/mysql/002_seed.sql` 种子数据 |

管理员账号由 `config.yaml` 的 `admin` 配置生成；生产环境请在 `deploy/backend.env` 中修改 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD`，不要沿用默认密码。
已有管理员不会在每次启动时自动重置密码；如需强制重置，临时设置 `ADMIN_RESET_PASSWORD_ON_STARTUP=true`。

也可以在用户端直接注册新账号。新注册用户默认 `0` 积分，这个默认值在后端代码中固定，不通过环境变量或配置文件覆盖。

## 配置

后端默认读取 `config.yaml`。Docker Compose 中通过下面的环境变量指定：

```text
CONFIG_PATH=/app/config.yaml
```

配置优先级：

1. 程序默认值
2. `config.yaml`
3. 环境变量

完整说明见 [config.yaml.md](config.yaml.md)。

生产环境重点修改：

- `auth.jwt_secret`：必须换成高强度随机字符串
- `admin.*`：默认管理员账号、密码、角色和状态
- `database.*`：数据库连接信息
- `storage.cos.*`：腾讯云 COS 存储桶、地域、公网域名
- `COS_SECRET_ID` / `COS_SECRET_KEY` 或 `SecretKey.csv`：腾讯云访问密钥

真实密钥不要提交到仓库。当前 `.gitignore` 已忽略：

- `SecretKey.csv`
- `*SecretKey*.csv`
- `*.secret.csv`
- `apikey`
- `url`

## 腾讯云 COS

平台上传的参考图片、参考视频、参考音频，以及生成结果会进入对象存储。

对象 key 会自动增加日期层，例如：

```text
20260628/references/1/image/1782625519125700420.jpg
20260628/videos/1/...
```

本地访问 `/api/assets/...` 时，如果请求来自 `localhost` 或 `127.0.0.1`，后端会转换为 COS 原始公网地址，方便上游 API 能访问参考文件。

## 上游 API 管理

管理员后台进入“上游 API 管理”后，可以：

- 新增上游 API：填写名称、Base URL、API Key、接口类型
- 同步上游模型
- 手动新增或编辑模型
- 启用或禁用 Provider、API Key、模型路由
- 管理图片模型和视频模型的上游关联

图片生成使用 OpenAI 兼容接口：

```text
POST /v1/images/generations
```

视频生成使用任务接口：

```text
POST /v1/videos
GET  /v1/videos/{task_id}
GET  /v1/videos/{task_id}/content
```

管理员只需要维护“平台模型 -> 上游 Provider + 上游模型名 + API Key”的关系，用户端选择模型后由后端自动路由。

## 常用开发命令

后端测试：

```bash
docker compose exec -T server go test ./...
```

前端类型检查：

```bash
docker compose exec -T user-web pnpm -r typecheck
```

用户端开发：

```bash
pnpm dev:user
```

管理员后台开发：

```bash
pnpm dev:admin
```

前端构建：

```bash
pnpm build
```

## 镜像构建与部署

GitHub Actions 文件：

```text
.github/workflows/docker-build-push.yml
```

触发方式：

- push 到 `main`
- push 到 `master`
- 手动 `workflow_dispatch`

会构建并推送后端镜像：

- `crpi-4otucz63tm2q5dhq.cn-beijing.personal.cr.aliyuncs.com/project-shiyu/agi-platform-server:latest`

需要在 GitHub 仓库配置 Secrets：

```text
ALIYUN_REGISTRY_USER
ALIYUN_REGISTRY_PASSWORD
```

服务器部署文件：

```text
docker-compose-deploy.yaml
docker-compose-deploy.yaml.md
deploy/backend.env.example
deploy/nginx-api.conf.example
```

服务器首次部署：

```bash
git clone <your-repo-url> agi-platform
cd agi-platform
cp deploy/backend.env.example deploy/backend.env
```

编辑 `deploy/backend.env`，至少修改：

```text
MYSQL_ROOT_PASSWORD
JWT_SECRET
ADMIN_USERNAME
ADMIN_PASSWORD
ADMIN_RESET_PASSWORD_ON_STARTUP
COS_SECRET_ID
COS_SECRET_KEY
COS_BUCKET
COS_REGION
COS_PUBLIC_BASE_URL
```

登录阿里云 ACR：

```bash
docker login crpi-4otucz63tm2q5dhq.cn-beijing.personal.cr.aliyuncs.com
```

启动后端和 MySQL：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d
```

检查状态：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml ps
curl http://127.0.0.1:18082/health
```

更新后端镜像：

```bash
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml pull server
docker compose --env-file deploy/backend.env -f docker-compose-deploy.yaml up -d server
```

更多重启、更新、日志命令见 [docker-compose-deploy.yaml.md](docker-compose-deploy.yaml.md)。

如果用户端和管理员后台部署到 Cloudflare Pages，服务器侧只需要使用后端镜像和 MySQL；两个前端项目分别在 Cloudflare 中配置 `VITE_API_BASE_URL=https://api.xxx.com`。`api.xxx.com` 可以用 nginx 或 Caddy 反向代理到 `http://127.0.0.1:18082`，nginx 示例见 `deploy/nginx-api.conf.example`。

## 数据库

初始化 SQL 位于：

```text
database/mysql/001_schema.sql
database/mysql/002_seed.sql
database/mysql/003_fix_video_reference_json.sql
```

开发环境首次启动 MySQL 时会自动执行 `database/mysql` 下的 SQL。

如果需要重建本地数据库：

```bash
docker compose down -v
docker compose up -d
```

## 相关文档

- [配置说明](config.yaml.md)
- [后端说明](server/README.md)
- [前端说明](web/README.md)
- [模块与数据库设计](docs/modules-and-database.md)
- [后台设计](docs/admin-design.md)
- [鉴权设计](docs/auth-design.md)
