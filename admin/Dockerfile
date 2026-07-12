# 管理后台前端
FROM node:18-alpine AS admin-builder

WORKDIR /app

# 复制 package.json
COPY package.json ./

# 安装 pnpm
RUN npm install -g pnpm

# 安装依赖
RUN pnpm install

# 复制源码
COPY . .

# 构建
RUN pnpm build

# 运行阶段 - 使用 nginx
FROM nginx:alpine

# 复制构建产物
COPY --from=admin-builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
