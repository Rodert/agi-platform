# 版本发布

`VERSION` 是项目发布版本的唯一来源，当前格式为 `主版本.次版本.补丁版本`。

日常功能修复和小更新只增加最后一位：

```text
0.1.0 -> 0.1.1 -> 0.1.2
```

发布步骤：

```bash
# 1. 修改 VERSION，例如写入 0.1.1
# 2. 提交并推送
git add VERSION
git commit -m "release: v0.1.1"
git push origin main

# 3. 为该提交创建版本标签并推送
git tag v0.1.1
git push origin v0.1.1
```

推送 `main` 后，GitHub Actions 会构建镜像并发布：

```text
ghcr.io/rodert/agi-platform:0.1.1
ghcr.io/rodert/agi-platform:latest
```

同时会创建 `v0.1.1` GitHub Release。后台版本管理会读取这些 Release，供更新和回滚使用。

不要只创建 Git tag 而不更新 `VERSION`，也不要复用已经发布过的版本号。
