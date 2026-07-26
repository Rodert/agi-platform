# 管理后台前端
FROM node:22-alpine AS admin-builder

WORKDIR /app

# 固定 pnpm 10，避免 Corepack 自动升级后拒绝 lockfile 中的构建脚本。
RUN corepack enable && corepack prepare pnpm@10.15.1 --activate

# 先复制依赖描述，以便构建缓存生效
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制源码
COPY . .

# 构建
RUN pnpm build

# 运行阶段 - 使用 nginx
FROM nginx:1.27-alpine

# 复制构建产物
COPY --from=admin-builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
