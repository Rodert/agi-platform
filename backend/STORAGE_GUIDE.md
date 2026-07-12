# 存储系统使用指南

## 📦 支持的存储类型

1. **本地存储** (local) - ✅ 已实现
2. **腾讯云 COS** (tencent_cos) - 待实现
3. **阿里云 OSS** (aliyun_oss) - 待实现  
4. **Cloudflare R2** (cloudflare) - 待实现

## 🚀 快速开始

### 1. 执行数据库迁移

```bash
mysql -u root -p agi_platform < scripts/migrations/002_add_storage_config.sql
```

### 2. 默认配置

系统会自动创建默认的本地存储配置：
- 存储路径: `./uploads`
- 访问域名: `http://localhost:8080/uploads`

### 3. 配置其他存储

通过管理后台或API添加其他存储配置。

## 📋 存储配置字段说明

| 字段 | 说明 | 本地存储 | 腾讯云COS | 阿里云OSS | Cloudflare R2 |
|------|------|----------|-----------|-----------|---------------|
| name | 配置名称 | ✅ | ✅ | ✅ | ✅ |
| type | 存储类型 | ✅ | ✅ | ✅ | ✅ |
| local_path | 本地路径 | ✅ | - | - | - |
| endpoint | 端点地址 | - | ✅ | ✅ | ✅ |
| access_key | AccessKey | - | ✅ | ✅ | ✅ |
| secret_key | SecretKey | - | ✅ | ✅ | ✅ |
| bucket | 桶名称 | - | ✅ | ✅ | ✅ |
| region | 区域 | - | ✅ | ✅ | - |
| domain | CDN域名 | ✅ | ✅ | ✅ | ✅ |

## 🔧 配置示例

### 本地存储

```json
{
  "name": "本地存储",
  "type": "local",
  "local_path": "./uploads",
  "domain": "http://localhost:8080/uploads"
}
```

### 腾讯云 COS

```json
{
  "name": "腾讯云COS",
  "type": "tencent_cos",
  "endpoint": "cos.ap-guangzhou.myqcloud.com",
  "access_key": "your_access_key",
  "secret_key": "your_secret_key",
  "bucket": "your-bucket",
  "region": "ap-guangzhou",
  "domain": "https://cdn.example.com"
}
```

### 阿里云 OSS

```json
{
  "name": "阿里云OSS",
  "type": "aliyun_oss",
  "endpoint": "oss-cn-hangzhou.aliyuncs.com",
  "access_key": "your_access_key",
  "secret_key": "your_secret_key",
  "bucket": "your-bucket",
  "domain": "https://cdn.example.com"
}
```

### Cloudflare R2

```json
{
  "name": "Cloudflare R2",
  "type": "cloudflare",
  "endpoint": "https://your-account-id.r2.cloudflarestorage.com",
  "access_key": "your_access_key",
  "secret_key": "your_secret_key",
  "bucket": "your-bucket",
  "domain": "https://cdn.example.com"
}
```

## 📝 使用说明

### 启用存储配置

同时只能启用一个存储配置。启用新配置时，会自动禁用其他配置。

### 切换存储

1. 在管理后台进入"存储配置"页面
2. 点击要启用的配置的"启用"按钮
3. 系统会自动切换到新的存储方式

### 文件上传

文件会自动上传到启用的存储后端，无需修改代码。

## 🔒 安全建议

1. **Secret Key 脱敏**: 前端显示时只显示后4位
2. **HTTPS**: 生产环境建议使用 HTTPS
3. **CDN**: 建议配置 CDN 域名加速访问
4. **权限控制**: 设置合适的存储桶权限

## 📦 依赖包（按需安装）

### 腾讯云 COS

```bash
go get github.com/tencentyun/cos-go-sdk-v5
```

### 阿里云 OSS

```bash
go get github.com/aliyun/aliyun-oss-go-sdk/oss
```

### Cloudflare R2

```bash
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
```

## 🐛 故障排查

### 本地存储权限问题

```bash
# 确保上传目录有写权限
chmod 755 ./uploads
```

### 云存储连接失败

1. 检查 AccessKey 和 SecretKey 是否正确
2. 检查 Endpoint 地址是否正确
3. 检查存储桶权限设置

## ⚙️ 迁移旧数据

如果之前使用 MinIO，可以：
1. 将 MinIO 的文件迁移到新的存储
2. 或保留 MinIO 作为 S3 兼容存储
3. 更新数据库中的文件 URL

