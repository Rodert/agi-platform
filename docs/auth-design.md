# 鉴权模块设计

## 一、目标

本阶段把临时 `user_id` 参数改成真实身份体系：

- 前台用户使用注册、登录、JWT 访问 `/api` 用户接口。
- 开发者使用 API Key 访问 `/v1` OpenAI 风格接口。
- Handler 只从请求上下文读取当前用户，不再信任请求体里的 `user_id`。
- Service 层不依赖 Gin，不关心 JWT 或 API Key 的解析细节。

## 二、接口

### 公开接口

- `POST /api/auth/register`
- `POST /api/auth/login`

### 用户接口

- `GET /api/me`
- `GET /api/api-keys`
- `POST /api/api-keys`
- `DELETE /api/api-keys/:id`
- `POST /api/images/generate`
- `GET /api/images/tasks/:task_no`

### 开发者接口

- `POST /v1/images/generations`

## 三、认证方式

### JWT

登录成功后返回：

```json
{
  "access_token": "xxx",
  "token_type": "Bearer",
  "expires_in": 604800
}
```

前台请求：

```http
Authorization: Bearer <access_token>
```

JWT claims：

```text
sub: user id
typ: user
exp: expire time
iat: issued time
```

### API Key

API Key 只在创建时返回一次，数据库只保存 hash 和 prefix。

格式：

```text
agi_<random>
```

开发者请求：

```http
Authorization: Bearer agi_xxx
```

或：

```http
X-API-Key: agi_xxx
```

## 四、上下文

中间件解析身份后写入 Gin context：

- `current_user_id`
- `current_api_key_id`
- `auth_type`

Handler 再把这些值传给 service。

## 五、低耦合约束

- JWT 签发与校验放在 `internal/auth`。
- Gin 中间件只负责 HTTP 解析与上下文写入。
- User/API Key 业务放在 service。
- 密码 hash、API Key hash 不暴露给 handler 外层。
- 生图 service 仍只接收 `UserID`、`APIKeyID`、`Source`，不感知具体认证方式。

