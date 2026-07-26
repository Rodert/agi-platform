# 📝 配置文件调整说明

## ✅ 已完成的调整

### 1. 移除配置文件中的动态配置

以下配置已从 `config.yaml` 和 `.env` 中移除，改为**从数据库读取**：

#### 已移除的配置项：
- ❌ `email.*` - 邮箱配置
- ❌ `payment.*` - 支付配置（易支付、支付宝、微信）
- ❌ `ai_models.*` - AI 模型配置

#### 保留的配置项：
- ✅ `server.*` - 服务器配置（端口、超时等）
- ✅ `database.*` - 数据库连接
- ✅ `redis.*` - Redis 连接
- ✅ `jwt.*` - JWT 密钥
- ✅ `storage.*` - 对象存储（本地存储 / Cloudflare R2）
- ✅ `system.*` - 系统配置（上传限制、CORS等）
- ✅ `worker.*` - Worker 配置

---

## 🗄️ 数据库存储的配置

### 1. 邮箱配置 - `email_config` 表

```sql
-- 单条记录，ID 固定为 1
id = 1
smtp_host = "smtp.gmail.com"
smtp_port = 587
smtp_user = "noreply@tide.ai"
smtp_password = "***"
smtp_ssl = false
from_name = "潮汐AI"
from_email = "noreply@tide.ai"
is_active = true
```

**管理接口**：
```
GET    /admin/v1/config/email      # 获取配置
PUT    /admin/v1/config/email      # 更新配置
POST   /admin/v1/config/email/test # 测试连接
```

---

### 2. 支付渠道配置 - `payment_channels` 表

```sql
-- 支持多个支付渠道
id = 1
name = "易支付-主账号"
channel_type = "epay"
merchant_id = "10001"
config = {
  "api_url": "https://pay.example.com",
  "partner_id": "10001",
  "partner_key": "xxx",
  "sign_type": "MD5",
  "notify_url": "https://api.tide.ai/webhook/epay"
}
is_active = true
sort_order = 1
```

**管理接口**：
```
GET    /admin/v1/payment/channels           # 列表
POST   /admin/v1/payment/channels           # 新增
PUT    /admin/v1/payment/channels/:id       # 编辑
PATCH  /admin/v1/payment/channels/:id/toggle # 启用/停用
DELETE /admin/v1/payment/channels/:id       # 删除
```

---

### 3. AI 模型配置 - `ai_models` 表

```sql
-- 支持多个 AI 模型
id = 1
name = "GPT Image2"
display_name = "GPT 图像 2 代"
type = "image"
provider = "openai"
cost = 4
api_config = {
  "api_url": "https://api.openai.com/v1/images/generations",
  "api_key": "sk-xxx",
  "model": "dall-e-3",
  "timeout": 120
}
params_config = {
  "ratio": {...},
  "resolution": {...}
}
is_active = true
sort_order = 1
```

**管理接口**：
```
GET    /admin/v1/ai-models              # 列表
POST   /admin/v1/ai-models              # 新增
PUT    /admin/v1/ai-models/:id          # 编辑
PATCH  /admin/v1/ai-models/:id/toggle   # 启用/停用
POST   /admin/v1/ai-models/:id/test     # 测试连接
```

---

## 💻 代码实现说明

### 1. 读取邮箱配置示例

```go
// internal/service/email.go
type EmailService struct {
    repo *repository.ConfigRepository
}

func (s *EmailService) SendVerificationCode(email, code string) error {
    // 从数据库读取邮箱配置
    emailConfig, err := s.repo.GetEmailConfig()
    if err != nil {
        return err
    }
    
    if !emailConfig.IsActive {
        return errors.New("邮箱服务未启用")
    }
    
    // 使用配置发送邮件
    dialer := mail.NewDialer(
        emailConfig.SMTPHost,
        emailConfig.SMTPPort,
        emailConfig.SMTPUser,
        emailConfig.SMTPPassword,
    )
    
    // ... 发送邮件逻辑
}
```

### 2. 读取支付渠道配置示例

```go
// internal/service/payment.go
type PaymentService struct {
    channelRepo *repository.PaymentChannelRepository
}

func (s *PaymentService) CreateOrder(userID int64, packageID int64) (*model.PaymentOrder, error) {
    // 获取启用的支付渠道
    channels, err := s.channelRepo.GetActiveChannels()
    if err != nil {
        return nil, err
    }
    
    // 选择支付渠道（这里选第一个，实际可以让用户选择）
    channel := channels[0]
    
    // 解析配置
    var config map[string]interface{}
    json.Unmarshal([]byte(channel.Config), &config)
    
    // 使用配置调用支付 API
    // ...
}
```

### 3. 读取 AI 模型配置示例

```go
// internal/service/creation.go
type CreationService struct {
    modelRepo *repository.AIModelRepository
}

func (s *CreationService) CreateTask(req *dto.CreateTaskRequest) error {
    // 从数据库读取模型配置
    aiModel, err := s.modelRepo.GetByName(req.ModelName)
    if err != nil {
        return errors.ErrNotFound
    }
    
    if !aiModel.IsActive {
        return errors.New("该模型已停用")
    }
    
    // 解析 API 配置
    var apiConfig map[string]interface{}
    json.Unmarshal([]byte(aiModel.APIConfig), &apiConfig)
    
    // 使用配置调用 AI API
    adapter := GetModelAdapter(aiModel.Provider)
    result := adapter.Generate(apiConfig, req.Prompt)
    
    // ...
}
```

---

## 🎯 优势

### 1. **动态配置**
- 无需重启服务即可修改配置
- 管理员可以在后台直接管理

### 2. **多租户支持**
- 支持多个支付渠道（多个易支付账号）
- 支持多个 AI 模型
- 可以随时启用/停用

### 3. **安全性**
- 敏感信息存储在数据库，不暴露在配置文件中
- 配置修改有操作日志（admin_logs 表）

### 4. **灵活性**
- 可以动态调整模型费用
- 可以动态调整模型参数
- 可以 A/B 测试不同的支付渠道

---

## 📋 待实现的功能

### 管理后台接口（Handler层）
- [ ] `internal/handler/admin/email_config.go`
- [ ] `internal/handler/admin/payment_channel.go`
- [ ] `internal/handler/admin/ai_model.go`

### Service 层
- [ ] `internal/service/email.go` - 从数据库读取配置
- [ ] `internal/service/payment.go` - 从数据库读取配置
- [ ] `internal/service/ai_model.go` - 从数据库读取配置

### Repository 层
- [ ] `internal/repository/config.go` - 配置相关查询
- [ ] `internal/repository/payment_channel.go` - 支付渠道查询
- [ ] `internal/repository/ai_model.go` - AI模型查询

---

## ✅ 总结

**配置策略**：
- **静态配置**（很少变动）→ 配置文件（server、database、redis、jwt、storage、system、worker）
- **动态配置**（经常变动）→ 数据库（email、payment、ai_models）

这样设计既保证了基础配置的稳定性，又提供了业务配置的灵活性！
