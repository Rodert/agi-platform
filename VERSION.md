# AGI-PLATFORM 发布指南

`VERSION` 是发布版本的唯一来源，格式为 `主版本.次版本.补丁版本`。

日常发布只递增最后一位：

```text
0.1.4 -> 0.1.5 -> 0.1.6
```

## 发布内容

每个版本都必须更新根目录 `VERSION` 和 `VERSION_RELEASE_NOTES.md`。

GitHub Release 标题固定为：

```text
AGI-PLATFORM v0.1.5
```

Release 内容使用以下结构，没有内容的栏目写“无”：

```md
## 新增功能

- 新增内容

## 优化改进

- 优化内容

## Bug 修复

- 修复内容
```

## 发布命令

以发布 `0.1.5` 为例：

```bash
# 1. 修改 VERSION 为 0.1.5，并更新 VERSION_RELEASE_NOTES.md
git add VERSION VERSION_RELEASE_NOTES.md
git commit -m "release: v0.1.5"
git push origin main

# 2. 为同一提交创建并推送标签
git tag v0.1.5
git push origin v0.1.5
```

推送 `main` 后，GitHub Actions 会构建并发布镜像：

```text
ghcr.io/rodert/agi-platform:0.1.5
ghcr.io/rodert/agi-platform:latest
ghcr.io/rodert/agi-platform:sha-<提交 SHA>
```

同时会创建 `AGI-PLATFORM v0.1.5` Release。后台版本管理以这些 Release 为更新和回滚版本来源。

## Docker 更新

手动拉取指定版本：

```bash
docker pull ghcr.io/rodert/agi-platform:0.1.5
```

在部署目录 `.env` 指定镜像后重启应用：

```env
APP_IMAGE=ghcr.io/rodert/agi-platform:0.1.5
```

```bash
docker compose pull app
docker compose up -d --no-deps app
```

使用 `latest` 跟随最新发布版本：

```env
APP_IMAGE=ghcr.io/rodert/agi-platform:latest
```

完整部署、GHCR 登录、后台一键更新和回滚配置见 [Docker 使用指南](./DOCKER_START.md)。

GitHub Actions 构建记录见：[Actions](https://github.com/Rodert/agi-platform/actions)。Release 列表见：[Releases](https://github.com/Rodert/agi-platform/releases)。

不要只创建 Git tag 而不更新 `VERSION`，也不要复用已经发布过的版本号。
