# config.yaml 中文说明

`config.yaml` 是后端服务的主要配置文件。Docker Compose 中通过 `CONFIG_PATH=/app/config.yaml` 指定读取路径。

配置优先级：

1. 程序默认值
2. `config.yaml`
3. 环境变量

也就是说，同名环境变量会覆盖 `config.yaml`。

## app

```yaml
app:
  env: local
  name: agi-platform
```

- `env`：运行环境标识，常用 `local`、`dev`、`prod`。
- `name`：应用名称，主要用于内部标识。

环境变量覆盖：

- `APP_ENV`
- `APP_NAME`

## http

```yaml
http:
  host: 0.0.0.0
  port: 8080
```

- `host`：后端监听地址。Docker 内通常用 `0.0.0.0`。
- `port`：后端监听端口。

环境变量覆盖：

- `HTTP_HOST`
- `HTTP_PORT`

## database

```yaml
database:
  host: mysql
  port: 3306
  user: root
  password: ""
  name: agi_platform
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_idle_conns: 10
  max_open_conns: 50
  conn_max_lifetime_seconds: 3600
```

- `host`：MySQL 地址。Docker Compose 内使用服务名 `mysql`。
- `port`：MySQL 端口。
- `user`：数据库用户名。
- `password`：数据库密码。
- `name`：数据库名。
- `charset`：字符集，建议保持 `utf8mb4`。
- `parse_time`：是否解析 MySQL 时间字段，建议保持 `true`。
- `loc`：时区位置，当前使用 `Local`。
- `max_idle_conns`：最大空闲连接数。
- `max_open_conns`：最大打开连接数。
- `conn_max_lifetime_seconds`：连接最大生命周期，单位秒。

环境变量覆盖：

- `MYSQL_HOST`
- `MYSQL_PORT`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_CHARSET`
- `MYSQL_PARSE_TIME`
- `MYSQL_LOC`
- `MYSQL_MAX_IDLE_CONNS`
- `MYSQL_MAX_OPEN_CONNS`
- `MYSQL_CONN_MAX_LIFETIME_SECONDS`

## auth

```yaml
auth:
  jwt_secret: local-dev-secret-change-me
  jwt_expire_seconds: 604800
```

- `jwt_secret`：JWT 签名密钥。生产环境必须改成高强度随机字符串。
- `jwt_expire_seconds`：登录 token 有效期，单位秒。`604800` 等于 7 天。
- 新注册用户默认 `0` 积分，这个值在后端代码中固定，不通过配置文件或环境变量覆盖。

环境变量覆盖：

- `JWT_SECRET`
- `JWT_EXPIRE_SECONDS`

## admin

```yaml
admin:
  enabled: true
  username: admin
  password: admin123
  reset_password_on_startup: false
  nickname: Administrator
  role: super_admin
  status: active
```

- `enabled`：是否在后端启动时自动确保默认管理员存在。
- `username`：默认管理员账号。
- `password`：默认管理员密码。首次创建管理员时必须至少 6 位。
- `reset_password_on_startup`：是否在每次启动时把已有管理员密码重置为 `password`。默认 `false`，避免覆盖管理员在个人中心修改后的密码。
- `nickname`：管理员昵称。
- `role`：管理员角色，默认 `super_admin`。
- `status`：管理员状态，默认 `active`。

环境变量覆盖：

- `ADMIN_BOOTSTRAP_ENABLED`
- `ADMIN_USERNAME`
- `ADMIN_PASSWORD`
- `ADMIN_RESET_PASSWORD_ON_STARTUP`
- `ADMIN_NICKNAME`
- `ADMIN_ROLE`
- `ADMIN_STATUS`

## storage

```yaml
storage:
  provider: cos
  local_root: uploads
  secret_csv_path: /app/SecretKey.csv
  cos:
    secret_id: ""
    secret_key: ""
    bucket: agi-platform-dev-1257142189
    region: ap-guangzhou
    public_base_url: https://agi-platform-dev-1257142189.cos.ap-guangzhou.myqcloud.com
    upload_prefix: ""
```

- `provider`：存储类型。当前支持：
  - `cos`：腾讯云 COS
  - `local`：本地文件，主要用于开发兜底
- `local_root`：本地存储目录，仅在显式设置 `provider: local` 时使用。
- `secret_csv_path`：腾讯云密钥 CSV 路径。当前 Docker 挂载后是 `/app/SecretKey.csv`。
- `cos.secret_id`：腾讯云 SecretId。建议不要写入仓库配置。
- `cos.secret_key`：腾讯云 SecretKey。建议不要写入仓库配置。
- `cos.bucket`：COS 存储桶名称。
- `cos.region`：COS 地域，例如广州是 `ap-guangzhou`。
- `cos.public_base_url`：COS 公网访问域名。
- `cos.upload_prefix`：上传前缀。为空时对象 key 直接从日期层开始，例如 `20260628/references/...`。

密钥读取顺序：

1. 环境变量 `COS_SECRET_ID`、`COS_SECRET_KEY`
2. `config.yaml` 中的 `cos.secret_id`、`cos.secret_key`
3. `secret_csv_path` 指向的 CSV 文件
4. 兜底查找 `../SecretKey.csv`、`SecretKey.csv`、`/app/SecretKey.csv`

环境变量覆盖：

- `STORAGE_PROVIDER`
- `STORAGE_LOCAL_ROOT`
- `COS_SECRET_CSV_PATH`
- `COS_SECRET_ID`
- `COS_SECRET_KEY`
- `COS_BUCKET`
- `COS_REGION`
- `COS_PUBLIC_BASE_URL`
- `COS_UPLOAD_PREFIX`

## 部署提醒

- 生产环境请修改 `auth.jwt_secret`。
- 生产环境请修改 `admin.password`，更推荐用环境变量 `ADMIN_PASSWORD` 配置，不要把真实管理员密码提交到仓库。需要强制重置管理员密码时，再临时设置 `ADMIN_RESET_PASSWORD_ON_STARTUP=true`。
- 生产环境建议保持 `storage.provider: cos`。COS 配置缺失或初始化失败时，服务会启动失败，不会自动退回写服务器本地磁盘。
- 不建议把真实 `cos.secret_id`、`cos.secret_key` 写进 `config.yaml` 并提交。
- 如果换 COS 桶，需要同步修改 `bucket`、`region`、`public_base_url`。
- 如果服务部署到自己的域名下，前端和数据库中仍保存 `/api/assets/...`，后端会负责跳转到 COS。
- 当前上传对象 key 会自动带日期层，例如 `20260628/images/...`、`20260628/videos/...`。
