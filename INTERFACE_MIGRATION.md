# 前后端接口协议统一改造文档

**改造日期**: 2026-07-12  
**改造原因**: 统一前后端接口协议，修复类型不匹配和路由不一致问题

---

## 📋 改造内容概览

### 1. 前端类型定义重构

#### 文件: `frontend/apps/web/src/types.ts`

**改造前问题**:
- ID 类型使用 `string`，后端返回 `number`
- 类型定义过于简化，缺少必要字段
- 枚举值使用中文，后端使用英文
- 缺少分页、模型等关键类型

**改造后**:
- ✅ 统一 ID 类型为 `number`
- ✅ 完善所有实体类型定义（User, Work, Task, AIModel 等）
- ✅ 添加分页响应类型 `PageResponse<T>`
- ✅ 统一枚举值为英文（MediaType, TaskStatus, AuditStatus 等）
- ✅ 字段名与后端 DTO 保持一致（snake_case）

**关键改动**:
```typescript
// 改造前
interface User { 
  name:string; avatar:string; level:string; 
  following:number; followers:number 
}

interface Work { 
  id:string;  // ❌ 类型错误
  author:string;  // ❌ 应该是对象
}

// 改造后
interface User {
  id: number
  email: string
  name: string
  avatar?: string
  bio?: string
  level: UserLevel
  invite_code: string
  following?: number
  followers?: number
  created_at: string
}

interface Work {
  id: number  // ✅ 正确类型
  user_id: number
  user?: User  // ✅ 对象类型
  title: string
  prompt: string
  category?: string
  type: MediaType
  image_url?: string
  video_url?: string
  likes_count: number
  collects_count: number
  is_liked: boolean
  is_collected: boolean
  created_at: string
}
```

---

### 2. API 客户端重构

#### 文件: `frontend/apps/web/src/utils/api.ts`

**改造内容**:
1. **统一响应结构**
   ```typescript
   interface ApiResponse<T = any> {
     success: boolean
     data?: T
     error?: { code: string; message: string }
     message?: string
   }
   ```

2. **完善 API 方法**
   - ✅ 添加商品图生成接口 `createProduct()`
   - ✅ 添加作品发布接口 `publish()`
   - ✅ 添加取消点赞/收藏接口 `unlike()`, `uncollect()`
   - ✅ 修正分页查询参数处理

3. **修复字段名不匹配**
   ```typescript
   // 改造前
   createImage: (data: {
     model_id: number,  // ❌ 后端使用 model_name
     width: number,     // ❌ 后端使用 params
     height: number
   })

   // 改造后
   createImage: (data: {
     model_name: string,  // ✅ 匹配后端
     params?: Record<string, any>,  // ✅ 通用参数
     reference_image?: string
   })
   ```

4. **完善请求处理**
   - 支持 FormData（文件上传场景）
   - 统一错误处理
   - 正确的参数序列化

---

### 3. 后端路由增强

#### 文件: `backend/cmd/api/main.go`

**改造内容**:

**问题**: 前端调用 `/works/:id/favorite`，但后端只实现了 `/works/:id/collect`

**解决方案**: 添加路由别名，同时支持两种路径

```go
// 改造前
works.POST("/:id/like", workHandler.LikeWork)
works.POST("/:id/favorite", workHandler.CollectWork)  // 缺少认证中间件

// 改造后
worksAuth := works.Group("")
worksAuth.Use(middleware.AuthMiddleware(&cfg.JWT))
{
    // 点赞
    worksAuth.POST("/:id/like", workHandler.LikeWork)
    worksAuth.DELETE("/:id/like", workHandler.UnlikeWork)

    // 收藏（同时支持两个路径）
    worksAuth.POST("/:id/collect", workHandler.CollectWork)
    worksAuth.DELETE("/:id/collect", workHandler.UncollectWork)
    worksAuth.POST("/:id/favorite", workHandler.CollectWork)      // 前端兼容
    worksAuth.DELETE("/:id/favorite", workHandler.UncollectWork)  // 前端兼容
}
```

**改进点**:
- ✅ 添加认证中间件（所有写操作需要登录）
- ✅ 支持 `favorite` 和 `collect` 两种路径
- ✅ 添加取消操作（DELETE 方法）
- ✅ 路由结构更清晰

---

## 🔍 数据库评估结果

### ✅ 无需修改数据库

经过对比分析，当前数据库表结构完善，已覆盖所有业务需求：

**已有表**:
- 用户模块: `users`, `verification_codes`
- 创作模块: `generation_requests`, `tasks`
- 作品模块: `works`, `work_likes`, `work_collects`, `work_audits`
- 积分模块: `credit_accounts`, `credit_ledgers`, `checkin_records`, `redeem_codes`, `credit_packages`
- 邀请模块: `invitations`, `invitation_rewards`
- 支付模块: `payment_orders`, `payment_channels`, `payment_transactions`
- 配置模块: `ai_models`, `system_configs`, `categories`, `email_config`, `storage_configs`
- 管理后台: `admin_users`, `admin_logs`

**评估结论**:
- 表结构设计合理，字段完整
- 索引配置正确
- 外键关系清晰
- 无需新增或修改表

**注意事项**:
- 前端定义的 `following`/`followers` 字段目前数据库没有对应的关注关系表
- 如果未来需要社交功能，可以新增 `user_follows` 表

---

## 📝 接口映射对照表

### 认证接口
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `auth.register()` | `POST /api/v1/auth/register` | ✅ |
| `auth.login()` | `POST /api/v1/auth/login` | ✅ |
| `auth.sendCode()` | `POST /api/v1/auth/send-code` | ✅ |

### 用户接口
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `user.getProfile()` | `GET /api/v1/users/profile` | ✅ |
| `user.updateProfile()` | `PATCH /api/v1/users/profile` | ✅ |

### 创作接口
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `generation.createImage()` | `POST /api/v1/generation/image` | ✅ |
| `generation.createVideo()` | `POST /api/v1/generation/video` | ✅ |
| `generation.createProduct()` | `POST /api/v1/generation/product` | 🆕 新增 |
| `generation.getModels()` | `GET /api/v1/generation/models` | ✅ |

### 任务接口
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `tasks.list()` | `GET /api/v1/tasks` | ✅ |
| `tasks.get()` | `GET /api/v1/tasks/:id` | ✅ |

### 作品接口
| 前端调用 | 后端路由 | 状态 |
|---------|---------|------|
| `works.list()` | `GET /api/v1/works` | ✅ |
| `works.get()` | `GET /api/v1/works/:id` | ✅ |
| `works.publish()` | `POST /api/v1/works` | 🆕 新增 |
| `works.like()` | `POST /api/v1/works/:id/like` | ✅ |
| `works.unlike()` | `DELETE /api/v1/works/:id/like` | 🆕 新增 |
| `works.collect()` | `POST /api/v1/works/:id/collect` | ✅ |
| `works.collect()` | `POST /api/v1/works/:id/favorite` | 🔧 路由别名 |
| `works.uncollect()` | `DELETE /api/v1/works/:id/collect` | 🆕 新增 |
| `works.uncollect()` | `DELETE /api/v1/works/:id/favorite` | 🔧 路由别名 |

---

## 🚀 部署说明

### 前端
1. 重新安装依赖（如有需要）:
   ```bash
   cd frontend
   pnpm install
   ```

2. 重新构建:
   ```bash
   pnpm build
   ```

3. 重启服务

### 后端
1. 重新编译:
   ```bash
   cd backend
   go build -o bin/api cmd/api/main.go
   ```

2. 重启服务:
   ```bash
   ./bin/api
   ```

**注意**: 数据库无需执行任何迁移脚本

---

## ✅ 兼容性说明

### 向后兼容
- ✅ 后端同时支持 `/favorite` 和 `/collect` 路径
- ✅ 旧代码可以正常工作
- ✅ 无需强制升级客户端

### 不兼容变更
- ⚠️ 前端 API 响应类型从 `string` ID 改为 `number` ID
- ⚠️ 需要前端代码同步更新，否则类型检查会报错

---

## 📊 改造影响范围

### 前端影响文件
- `frontend/apps/web/src/types.ts` - 类型定义（完全重写）
- `frontend/apps/web/src/utils/api.ts` - API 客户端（完全重写）
- 其他使用这些类型的组件（需要根据新类型调整）

### 后端影响文件
- `backend/cmd/api/main.go` - 路由配置（局部修改）
- 其他文件无需修改

### 数据库影响
- 无

---

## 🧪 测试建议

### 接口测试
```bash
# 测试认证
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password","type":"password"}'

# 测试作品收藏（两种路径）
curl -X POST http://localhost:8080/api/v1/works/1/collect \
  -H "Authorization: Bearer YOUR_TOKEN"

curl -X POST http://localhost:8080/api/v1/works/1/favorite \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 类型检查
```bash
cd frontend
pnpm type-check
```

---

## 📌 后续优化建议

1. **社交功能**
   - 如需实现关注/粉丝功能，新增 `user_follows` 表
   - 实现相关 API 接口

2. **API 版本管理**
   - 考虑引入 API 版本控制（v1, v2）
   - 避免破坏性变更影响客户端

3. **类型生成**
   - 考虑使用工具从后端自动生成前端类型定义
   - 推荐: OpenAPI + TypeScript Codegen

4. **文档**
   - 使用 Swagger/OpenAPI 生成 API 文档
   - 保持前后端文档同步

---

**改造完成** ✅
