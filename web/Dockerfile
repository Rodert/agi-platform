FROM node:22-alpine AS builder

WORKDIR /app

ARG APP_NAME=user
ARG VITE_API_BASE_URL=http://127.0.0.1:8080

ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml tsconfig.base.json ./
COPY web ./web

RUN corepack enable \
  && pnpm config set store-dir /pnpm/store \
  && pnpm install --frozen-lockfile \
  && pnpm --filter @agi-platform/${APP_NAME} build

FROM nginx:1.27-alpine

ARG APP_NAME=user

COPY web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/web/${APP_NAME}/dist /usr/share/nginx/html

EXPOSE 80
