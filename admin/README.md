# AGI Platform Admin - 管理后台

基于 Vue 3 + Vite + Element Plus 的管理后台前端。

## 功能特性

- ✅ 管理员登录
- ✅ 数据统计、用户与权限管理
- ✅ 作品审核、任务与日志管理
- ✅ AI 模型、渠道、邮件、充值套餐与兑换码配置
- ✅ 版本检查、更新与回滚入口
- ✅ 响应式设计

## 快速开始

### 安装依赖

```bash
pnpm install
```

### 启动开发服务器

```bash
pnpm dev
```

访问：http://localhost:3001

### 构建生产版本

```bash
pnpm build
```

## 超级管理员

首次启动时由环境变量 `SUPER_ADMIN_USERNAME`、`SUPER_ADMIN_PASSWORD` 和 `SUPER_ADMIN_NAME` 创建。部署与本地调试配置见根目录 [DEVELOP.md](../DEVELOP.md)。

## 技术栈

- Vue 3
- Vite
- Vue Router
- Pinia
- Element Plus
- Axios
- Sass

## 目录结构

```
admin/
├── src/
│   ├── api/          # API 接口
│   ├── layout/       # 布局组件
│   ├── router/       # 路由配置
│   ├── stores/       # 状态管理
│   ├── styles/       # 全局样式
│   ├── utils/        # 工具函数
│   ├── views/        # 页面组件
│   ├── App.vue       # 根组件
│   └── main.js       # 入口文件
├── index.html
├── package.json
└── vite.config.js
```

## 接口对接

开发服务器默认运行在 `http://localhost:3001`，后端 API 通过代理转发到 `http://localhost:8080`；使用 Docker 本地调试时，请从统一入口 `http://localhost:3100/admin/` 访问。

所有管理后台接口前缀为 `/admin/v1`
