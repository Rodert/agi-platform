import type { User, Work, Task, AIModel, Announcement, Ledger, PageResponse, UserSession, CreditPackage } from '../types'

// Production requests stay on the current origin so the reverse proxy handles
// `/api`. Ignore a stale local-development URL that could otherwise be baked
// into a production Docker image.
const configuredApiBaseURL = import.meta.env.VITE_API_BASE_URL?.trim() || ''
const API_BASE_URL = /^https?:\/\/(?:localhost|127\.0\.0\.1)(?::\d+)?(?:\/|$)/i.test(configuredApiBaseURL)
  ? ''
  : configuredApiBaseURL

// 统一响应结构
interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: {
    code: number | string
    message: string
  }
  message?: string
}

export class ApiError extends Error {
  constructor(message: string, public readonly code?: number | string) {
    super(message)
    this.name = 'ApiError'
  }
}

// API客户端
class APIClient {
  private baseURL: string
  private token: string | null = null

  constructor(baseURL: string) {
    this.baseURL = baseURL
    this.token = localStorage.getItem('token')
  }

  setToken(token: string | null) {
    this.token = token
    if (token) {
      localStorage.setItem('token', token)
    } else {
      localStorage.removeItem('token')
    }
  }

  getToken() {
    return this.token
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers = new Headers(options.headers)

    // 如果不是 FormData，设置 Content-Type
    if (!(options.body instanceof FormData)) {
      headers.set('Content-Type', 'application/json')
    }

    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    })

    const data: ApiResponse<T> = await response.json()

    if (!response.ok || !data.success) {
      throw new ApiError(data.error?.message || data.message || '请求失败', data.error?.code)
    }

    return data.data as T
  }

  // 用户认证API
  auth = {
    // 注册
    register: (data: {
      email: string
      password: string
      confirm_password: string
      code?: string
      invite_code?: string
    }) =>
      this.request<{ token: string; user: User }>(
        '/api/v1/auth/register',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    // 登录
    login: (data: {
      email: string
      password?: string
      code?: string
      type: 'password' | 'code'
    }) =>
      this.request<{ token: string; user: User }>(
        '/api/v1/auth/login',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    // 发送验证码
    sendCode: (data: { email: string; type: 'register' | 'login' | 'reset' }) =>
      this.request<void>(
        '/api/v1/auth/send-code',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    registrationSettings: () => this.request<{ register_email_verification: boolean }>('/api/v1/auth/registration-settings'),
  }

  // 用户API
  user = {
    // 获取个人信息
    getProfile: () =>
      this.request<User>('/api/v1/users/profile'),

    // 更新个人信息
    updateProfile: (data: { name?: string; avatar?: string; bio?: string }) =>
      this.request<User>(
        '/api/v1/users/profile',
        { method: 'PATCH', body: JSON.stringify(data) }
      ),

    bindPhone: (data: { phone: string; code: string }) =>
      this.request<User>('/api/v1/users/phone', { method: 'POST', body: JSON.stringify(data) }),

    changePassword: (data: { current_password: string; new_password: string }) =>
      this.request<void>('/api/v1/users/password', { method: 'POST', body: JSON.stringify(data) }),

    listSessions: () => this.request<UserSession[]>('/api/v1/users/sessions'),

    revokeSession: (id: string) => this.request<void>(`/api/v1/users/sessions/${id}`, { method: 'DELETE' }),

    getCreditLedgers: (params: { page?: number; page_size?: number } = {}) => {
      const query = new URLSearchParams()
      if (params.page) query.set('page', String(params.page))
      if (params.page_size) query.set('page_size', String(params.page_size))
      return this.request<PageResponse<Ledger>>(`/api/v1/users/credits?${query.toString()}`)
    },

    redeemCode: (code: string) => this.request<{ amount: number; balance: number }>('/api/v1/users/redeem-codes', {
      method: 'POST', body: JSON.stringify({ code }),
    }),
  }

  creditPackages = () => this.request<CreditPackage[]>('/api/v1/credit-packages')

  // 创作API
  generation = {
    // 创建图片任务
    createImage: (data: {
      prompt: string
      model_name: string
      params?: Record<string, any>
      reference_image?: string
    }) =>
      this.request<Task>(
        '/api/v1/generation/image',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    // 创建视频任务
    createVideo: (data: {
      prompt: string
      model_name: string
      params?: Record<string, any>
      first_frame_url?: string
      last_frame_url?: string
      reference_images?: string[]
    }) =>
      this.request<Task>(
        '/api/v1/generation/video',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    optimizePrompt: (data: {
      prompt: string
      target_type: 'image' | 'video'
      target_model_name?: string
      params?: Record<string, any>
    }) =>
      this.request<{ prompt: string; model_name: string; credit_cost: number }>(
        '/api/v1/generation/prompt-optimization',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    // 获取模型列表
    getModels: (type?: 'image' | 'video') => {
      const params = type ? `?type=${type}` : ''
      return this.request<AIModel[]>(`/api/v1/generation/models${params}`)
    },
  }

  // 任务API
  tasks = {
    // 获取任务列表
    list: (params: {
      page?: number
      page_size?: number
      status?: string
      type?: string
    }) => {
      const query = new URLSearchParams()
      if (params.page) query.set('page', String(params.page))
      if (params.page_size) query.set('page_size', String(params.page_size))
      if (params.status) query.set('status', params.status)
      if (params.type) query.set('type', params.type)

      return this.request<PageResponse<Task>>(
        `/api/v1/tasks?${query.toString()}`
      )
    },

    // 获取任务详情
    get: (id: number) =>
      this.request<Task>(`/api/v1/tasks/${id}`),

    // 下载任务结果，使用受鉴权保护的流式接口，兼容私有对象存储。
    download: async (id: number) => {
      const response = await fetch(`${this.baseURL}/api/v1/tasks/${id}/download`, {
        headers: this.token ? { Authorization: `Bearer ${this.token}` } : {},
      })
      if (!response.ok) {
        const data = await response.json().catch(() => null) as ApiResponse<unknown> | null
        throw new Error(data?.error?.message || data?.message || '下载失败')
      }
      const blob = await response.blob()
      const filename = response.headers.get('content-disposition')?.match(/filename=([^;]+)/i)?.[1] || `agi-task-${id}`
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    },
  }

  notifications = {
    list: (params: { page?: number; page_size?: number } = {}) => {
      const query = new URLSearchParams()
      if (params.page) query.set('page', String(params.page))
      if (params.page_size) query.set('page_size', String(params.page_size))
      return this.request<PageResponse<Announcement>>(`/api/v1/announcements?${query.toString()}`)
    },
  }

  // 作品API
  works = {
    // 获取作品列表
    list: (params: {
      page?: number
      page_size?: number
      category?: string
      type?: string
      user_id?: number
    }) => {
      const query = new URLSearchParams()
      if (params.page) query.set('page', String(params.page))
      if (params.page_size) query.set('page_size', String(params.page_size))
      if (params.category) query.set('category', params.category)
      if (params.type) query.set('type', params.type)
      if (params.user_id) query.set('user_id', String(params.user_id))

      return this.request<PageResponse<Work>>(
        `/api/v1/works?${query.toString()}`
      )
    },

    mine: (params: { page?: number; page_size?: number } = {}) => {
      const query = new URLSearchParams()
      if (params.page) query.set('page', String(params.page))
      if (params.page_size) query.set('page_size', String(params.page_size))
      return this.request<PageResponse<Work>>(`/api/v1/works/mine?${query.toString()}`)
    },

    // 获取作品详情
    get: (id: number) =>
      this.request<Work>(`/api/v1/works/${id}`),

    // 发布作品
    publish: (data: { task_id: number; title: string; category?: string }) =>
      this.request<Work>(
        '/api/v1/works',
        { method: 'POST', body: JSON.stringify(data) }
      ),

    // 点赞作品
    like: (id: number) =>
      this.request<void>(
        `/api/v1/works/${id}/like`,
        { method: 'POST' }
      ),

    // 取消点赞
    unlike: (id: number) =>
      this.request<void>(
        `/api/v1/works/${id}/like`,
        { method: 'DELETE' }
      ),

    // 收藏作品
    collect: (id: number) =>
      this.request<void>(
        `/api/v1/works/${id}/collect`,
        { method: 'POST' }
      ),

    // 取消收藏
    uncollect: (id: number) =>
      this.request<void>(
        `/api/v1/works/${id}/collect`,
        { method: 'DELETE' }
      ),
  }
}

// 导出单例
export const apiClient = new APIClient(API_BASE_URL)

// 导出类型
export type { ApiResponse }
