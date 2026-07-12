# AGI Platform Admin - 管理后台

基于 Vue 3 + Vite + Element Plus 的管理后台前端。

## 功能特性

- ✅ 管理员登录
- ✅ 数据统计看板
- ✅ 作品审核（通过/拒绝）
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

## 默认账号

- 用户名：`admin`
- 密码：`admin123`

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

后端API地址通过代理配置到 `http://localhost:8080`

所有管理后台接口前缀为 `/admin/v1`
