# 后台管理模块设计

## 一、目标

本阶段搭建后台最小可运营闭环：

- 管理员登录与后台 JWT。
- 后台查看用户列表、手动加减积分。
- 后台管理 Provider。
- 后台管理上游 Provider Key。
- 后台管理前台图片模型和模型路由。

## 二、接口边界

### 公开后台接口

- `POST /admin/auth/login`

### 需要管理员登录

- `GET /admin/me`
- `GET /admin/users`
- `POST /admin/users/:id/credits`
- `GET /admin/providers`
- `POST /admin/providers`
- `PUT /admin/providers/:id`
- `GET /admin/providers/:id/keys`
- `POST /admin/providers/:id/keys`
- `DELETE /admin/provider-keys/:id`
- `GET /admin/image-models`
- `POST /admin/image-models`
- `PUT /admin/image-models/:id`
- `GET /admin/image-models/:id/routes`
- `POST /admin/image-models/:id/routes`
- `PUT /admin/image-model-routes/:id`
- `GET /admin/image-tasks`

## 三、鉴权设计

后台和前台都使用 JWT，但 claims 中的 `typ` 不同：

- 前台用户：`typ=user`
- 后台管理员：`typ=admin`

后台中间件只接受 `typ=admin` 的 token。

## 四、低耦合约束

- Handler 只做参数绑定和响应转换。
- Service 负责业务规则，例如手动加减积分必须写钱包流水。
- Repository 负责 GORM 访问。
- 管理员操作日志独立成 repository，后续可在 service 中逐步补齐审计。
- Provider Key 暂时按原字段写入 `api_key_encrypted`，后续可替换为独立密钥加密组件，不影响 service 接口。

## 五、第一版暂不做

- 细粒度 RBAC 权限。
- 操作日志全量覆盖。
- Provider Key 真加密和脱敏展示策略细化。
- 后台分页总数统计。

