import type { User, Work, Task, AIModel, PageResponse } from '../types'

// API配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

// 统一响应结构
interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
  message?: string
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
      throw new Error(data.error?.message || data.message || '请求失败')
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
      code: string
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
  }

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
    }) =>
      this.request<Task>(
        '/api/v1/generation/video',
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
