# 更新日志

## [未发布] - 存储系统重构

### ✨ 新增
- 支持多种存储后端（本地、腾讯云COS、阿里云OSS、Cloudflare R2）
- 存储配置可在数据库动态管理
- 支持在管理后台启用/停用存储配置
- Secret Key 自动脱敏显示
- 支持自定义 CDN 域名

### 🔄 变更
- 移除旧对象存储依赖，简化部署
- 默认使用本地存储
- Docker 配置已更新

### 📦 依赖
- 新增（可选）：
  - 腾讯云 COS SDK
  - 阿里云 OSS SDK
  - AWS SDK (用于 Cloudflare R2)

### 📝 迁移指南
1. 执行数据库迁移：
   ```bash
   mysql -u root -p agi_platform < backend/scripts/migrations/002_add_storage_config.sql
   ```

2. 更新 Docker 配置：
   ```bash
   docker-compose down
   docker-compose up -d --build
   ```

3. 默认使用本地存储，无需额外配置

4. 如需使用云存储：
   - 在管理后台添加存储配置
   - 点击"启用"按钮切换存储

---

## [v1.0.0] - 初始版本

### ✨ 功能
- 用户认证系统
- AI 创作任务管理
- Worker 异步处理
- 作品社区功能
- 管理后台
- Docker 一键部署
