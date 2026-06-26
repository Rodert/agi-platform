# AGI Platform 完整产品路线

## 一、完整模块

AGI Platform 最终需要完成以下模块：

1. 用户端 Web
   - 登录注册
   - AI 生图工作台
   - 我的作品
   - API Key 管理
   - 充值中心
   - API 接入文档

2. 管理后台 Web
   - 管理员登录
   - 用户管理
   - 手动加减积分
   - Provider 管理
   - Provider Key 管理
   - 图片模型管理
   - 模型路由管理
   - 生成任务管理
   - 订单管理
   - 系统配置

3. 后端核心
   - 用户鉴权
   - 管理员鉴权
   - API Key 鉴权
   - 图片生成任务
   - 积分扣费与退款
   - Provider 适配层
   - 图片存储层
   - 支付订单层
   - 审计日志

4. 上游模型适配
   - Mock，本地开发
   - OpenAI GPT Image
   - xAI/Grok Image
   - Gemini/Nano Banana
   - OpenRouter
   - OpenAI-compatible 自定义 Provider

5. 存储与支付
   - 本地图片存储
   - Cloudflare R2 / OSS / COS / MinIO
   - 手动充值
   - 支付回调
   - 套餐配置

## 二、本轮交付目标

本轮先完成前端可用 MVP：

- `web/user`：用户端应用
- `web/admin`：管理后台应用
- `web/shared`：共享 API Client 与类型

用户端先接入：

- 注册 / 登录
- 当前用户余额
- 模型列表
- 生图提交
- 任务结果展示
- API Key 创建和列表

管理端先接入：

- 管理员登录
- 用户列表
- 手动加减积分
- Provider 管理
- Provider Key 管理
- 图片模型管理
- 模型路由管理
- 生成任务列表

## 三、低耦合约束

- 前端共享请求逻辑放在 `web/shared`。
- 用户端和管理端不直接互相依赖。
- 每个应用只维护自己的页面、状态和路由。
- API URL、token key、应用配置集中管理。
- 后端 service/handler/repository 分层继续保持。

