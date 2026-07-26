## 新增功能

- GitHub Container Registry 镜像同时发布 `linux/amd64` 与 `linux/arm64` 架构。
- 后端版本检查支持可选的 `GITHUB_TOKEN`，可使用认证请求提高 GitHub API 额度。

## 优化改进

- 版本检查改由后端访问 GitHub，并通过 Redis 缓存一小时；打开版本管理默认读取缓存。
- Docker 构建和部署文档补充多架构镜像与版本检查配置说明。

## Bug 修复

- 修复后台页面刷新时直接请求 GitHub API 导致频繁触发匿名限流的问题。
- GitHub API 请求失败或限流时也会缓存一小时，避免管理端重复请求。
