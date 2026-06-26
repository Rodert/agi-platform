# AGI Platform 模块与数据库设计

## 一、核心模块

AGI Platform 第一版建议拆成以下模块：

1. 用户与鉴权模块
   - 用户注册、登录、状态管理
   - 管理员账号与后台权限
   - 开发者 API Key

2. 钱包与积分模块
   - 用户积分余额
   - 积分消费、退款、后台加减积分
   - 积分流水审计

3. 模型与 Provider 模块
   - 前台展示模型
   - 上游 Provider 配置
   - 多 API Key 轮询
   - 模型路由、价格、支持尺寸

4. 图片生成任务模块
   - 文生图任务
   - 任务状态流转
   - 上游请求与响应记录
   - 失败重试、失败退款

5. 作品与存储模块
   - 生成图片记录
   - 图片转存后的对象存储地址
   - 下载、删除、违规标记

6. 订单与支付模块
   - 积分套餐
   - 充值订单
   - 支付回调记录
   - 后台补单

7. 开放 API 模块
   - OpenAI 风格图片生成接口
   - API Key 鉴权
   - API 调用日志与用量统计

8. 系统配置与运营模块
   - 网站基础配置
   - 注册赠送积分
   - 默认模型
   - 敏感词、公告、客服信息

9. 审计日志模块
   - 后台操作日志
   - 钱包变动留痕
   - 任务异常排查

## 二、表清单

### 用户与鉴权

- `users`：前台用户
- `admin_users`：后台管理员
- `api_keys`：开发者 API Key

### 钱包与积分

- `wallet_logs`：积分流水

### 模型与 Provider

- `providers`：上游服务商
- `provider_keys`：上游 API Key 池
- `image_models`：前台可选模型
- `image_model_routes`：模型到上游 Provider 的路由配置

### 图片生成与作品

- `image_tasks`：生成任务
- `image_assets`：生成后的图片作品

### 订单与支付

- `credit_packages`：积分套餐
- `orders`：充值订单
- `payment_callbacks`：支付回调日志

### API 与日志

- `api_request_logs`：开放 API 调用日志
- `admin_operation_logs`：后台操作日志

### 系统配置

- `system_settings`：系统配置项

## 三、设计原则

- 金额使用 `DECIMAL(10,2)`，积分使用整数 `BIGINT`。
- 上游响应、支持尺寸、扩展参数使用 MySQL `JSON` 字段，避免第一版过度拆表。
- 用户余额冗余在 `users.credits`，每次变动必须写入 `wallet_logs`。
- 图片任务和图片作品分开：一个任务可以生成多张图片。
- 前台模型和上游真实模型分开：`image_models` 面向用户，`image_model_routes` 面向 Provider。
- 上游 API Key 单独建表，方便做轮询、限额、禁用和失败隔离。
- 所有核心表保留 `created_at`、`updated_at`，需要软删除的表使用 `deleted_at`。

