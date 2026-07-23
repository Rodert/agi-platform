// 媒体类型
export type MediaType = 'image' | 'video'

// 任务状态
export type TaskStatus = 'queued' | 'processing' | 'success' | 'failed'

// 审核状态
export type AuditStatus = 'pending' | 'approved' | 'rejected'

// 用户等级
export type UserLevel = 'free' | 'member' | 'pro'

// 积分类型
export type LedgerType = 'income' | 'expense'

// 用户信息
export interface User {
  id: number
  email: string
  name: string
  avatar?: string
  bio?: string
  level: UserLevel
  balance: number
  invite_code: string
  following?: number
  followers?: number
  created_at: string
}

// 用户资料（完整）
export interface UserProfile extends User {
  // 可以扩展其他字段
}

// 作品
export interface Work {
  id: number
  user_id: number
  user?: User
  title: string
  prompt: string
  category?: string
  type: MediaType
  ratio?: string
  image_url?: string
  video_url?: string
  audit_status: AuditStatus
  likes_count: number
  collects_count: number
  views_count: number
  is_liked: boolean
  is_collected: boolean
  published_at?: string
  created_at: string
}

// 任务
export interface Task {
  id: number
  title: string
  type: MediaType
  status: TaskStatus
  progress: number
  prompt: string
  model_name: string
  result_url?: string
  thumbnail_url?: string
  error_msg?: string
  cost: number
  created_at: string
  completed_at?: string
}

// 模型
export interface AIModel {
  id: number
  name: string
  display_name: string
  type: MediaType
  provider: string
  description?: string
  logo_url?: string
  tag?: string
  cost: number
  params_config?: Record<string, ModelParamConfig>
}

export interface ModelParamOption {
  value: string
  label: string
  extra_cost?: number
}

export interface ModelParamConfig {
  label: string
  type: 'select' | 'switch'
  default?: string | boolean
  options?: ModelParamOption[]
}

// 积分流水
export interface Ledger {
  id: number
  title: string
  amount: number
  type: LedgerType
  balance_after: number
  created_at: string
}

export interface Announcement {
  id: number
  title: string
  content: string
  category: string
  published_at?: string
  created_at: string
}

// 分页响应
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
