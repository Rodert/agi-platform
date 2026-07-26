FROM golang:1.23-alpine AS backend-builder

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api cmd/api/main.go \
    && CGO_ENABLED=0 GOOS=linux go build -o /out/worker cmd/worker/main.go

FROM node:22-alpine AS frontend-builder

WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@10.15.1 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
COPY frontend/apps/web/package.json ./apps/web/package.json
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build:web

FROM node:22-alpine AS admin-builder

WORKDIR /src/admin
RUN corepack enable && corepack prepare pnpm@10.15.1 --activate
COPY VERSION /src/VERSION
COPY admin/package.json admin/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY admin/ ./
RUN pnpm build

FROM alpine:3.21

RUN apk add --no-cache ca-certificates nginx supervisor tzdata wget \
    && mkdir -p /app/uploads /run/nginx /var/log/supervisor

WORKDIR /app
ENV TZ=Asia/Shanghai

COPY --from=backend-builder /out/api /app/api
COPY --from=backend-builder /out/worker /app/worker
COPY --from=backend-builder /src/backend/configs /app/configs
COPY --from=frontend-builder /src/frontend/apps/web/dist /usr/share/nginx/html
COPY --from=admin-builder /src/admin/dist /usr/share/nginx/admin
COPY deploy/nginx.conf /etc/nginx/http.d/default.conf
COPY deploy/supervisord.conf /etc/supervisord.conf

EXPOSE 80
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
