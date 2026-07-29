import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import type { AIModel, Ledger, Task, User, Work } from '../types'
import { CREDIT_PACKAGES_EVENT, requestLogin } from '../utils/auth'
import { ApiError, apiClient } from '../utils/api'

interface CreateTaskInput {
  prompt: string
  modelName: string
  type: Task['type']
  params?: Record<string, unknown>
  referenceImage?: string
  referenceImages?: string[]
  firstFrameUrl?: string
  lastFrameUrl?: string
}

type CreateTaskResult = 'success' | 'insufficient_credit' | 'failed'

interface Store {
  user: User | null
  authReady: boolean
  balance: number
  works: Work[]
  tasks: Task[]
  models: AIModel[]
  ledger: Ledger[]
  authenticate: (input: { email: string; password?: string; code?: string; type: 'password' | 'code' }) => Promise<boolean>
  register: (input: { email: string; password: string; confirm_password: string; code?: string }) => Promise<boolean>
  logout: () => void
  requireAuth: () => boolean
  toggleLike: (id: number) => Promise<void>
  toggleCollect: (id: number) => Promise<void>
  createTask: (input: CreateTaskInput) => Promise<CreateTaskResult>
  retryTask: (id: number) => void
  recharge: () => void
  checkIn: () => boolean
  redeem: () => boolean
  loadTasks: () => Promise<void>
	loadMoreTasks: () => Promise<void>
	hasMoreTasks: boolean
	loadingMoreTasks: boolean
	loadWorks: () => Promise<void>
	refreshProfile: () => Promise<void>
}

const Context = createContext<Store | null>(null)

export function AppProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [authReady, setAuthReady] = useState(false)
  const [works, setWorks] = useState<Work[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [taskTotal, setTaskTotal] = useState(0)
  const [loadingMoreTasks, setLoadingMoreTasks] = useState(false)
  const nextTaskPageRef = useRef(2)
  const [models, setModels] = useState<AIModel[]>([])

  const mergeTasks = useCallback((current: Task[], incoming: Task[]) => {
    const merged = new Map(current.map(task => [task.id, task]))
    incoming.forEach(task => merged.set(task.id, task))
    return [...merged.values()].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }, [])

  const loadUserProfile = useCallback(async () => {
    try {
      setUser(await apiClient.user.getProfile())
    } catch (error) {
      console.error('加载用户信息失败:', error)
      apiClient.setToken(null)
      setUser(null)
    }
  }, [])

  const loadTasks = useCallback(async () => {
    if (!apiClient.getToken()) return
    try {
      const result = await apiClient.tasks.list({ page: 1, page_size: 50 })
      setTaskTotal(result.total)
      setTasks(current => mergeTasks(current, result.list))
    } catch (error) {
      console.error('加载任务失败:', error)
    }
  }, [mergeTasks])

  const loadMoreTasks = useCallback(async () => {
    if (!apiClient.getToken() || loadingMoreTasks || tasks.length >= taskTotal) return
    setLoadingMoreTasks(true)
    try {
      const result = await apiClient.tasks.list({ page: nextTaskPageRef.current, page_size: 50 })
      setTaskTotal(result.total)
      setTasks(current => mergeTasks(current, result.list))
      nextTaskPageRef.current += 1
    } catch (error) {
      console.error('加载更多任务失败:', error)
    } finally {
      setLoadingMoreTasks(false)
    }
  }, [loadingMoreTasks, mergeTasks, taskTotal, tasks.length])

  const loadWorks = useCallback(async () => {
    try {
      const result = await apiClient.works.list({ page: 1, page_size: 50 })
      setWorks(result.list)
    } catch (error) {
      console.error('加载作品失败:', error)
    }
  }, [])

  useEffect(() => {
    const initialize = async () => {
      await Promise.all([
        loadWorks(),
        apiClient.generation.getModels().then(setModels).catch(error => console.error('加载模型失败:', error)),
        apiClient.getToken() ? Promise.all([loadUserProfile(), loadTasks()]) : Promise.resolve(),
      ])
      setAuthReady(true)
    }
    void initialize()
  }, [loadTasks, loadUserProfile, loadWorks])

  useEffect(() => {
    if (!user || !tasks.some(task => ['queued', 'processing', 'uploading'].includes(task.status))) return
    const timer = window.setInterval(() => { void loadTasks() }, 3000)
    return () => window.clearInterval(timer)
  }, [loadTasks, tasks, user])

  const completeAuthentication = async (result: { token: string }) => {
    apiClient.setToken(result.token)
    setAuthReady(false)
    await Promise.all([loadUserProfile(), loadTasks(), loadWorks()])
    setAuthReady(true)
  }

  const authenticate = async (input: { email: string; password?: string; code?: string; type: 'password' | 'code' }) => {
    try {
      const result = await apiClient.auth.login(input)
      await completeAuthentication(result)
      return true
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  const register = async (input: { email: string; password: string; confirm_password: string; code?: string }) => {
    try {
      const result = await apiClient.auth.register(input)
      await completeAuthentication(result)
      return true
    } catch (error) {
      console.error('注册失败:', error)
      return false
    }
  }

  const logout = () => {
    apiClient.setToken(null)
    setUser(null)
    setTasks([])
    setTaskTotal(0)
    nextTaskPageRef.current = 2
    setAuthReady(true)
  }

  const requireAuth = () => {
    if (user) return true
    requestLogin()
    return false
  }

  const toggleLike = async (id: number) => {
    if (!requireAuth()) return
    const work = works.find(item => item.id === id)
    if (!work) return
    try {
      await (work.is_liked ? apiClient.works.unlike(id) : apiClient.works.like(id))
      setWorks(items => items.map(item => item.id === id ? {
        ...item,
        is_liked: !item.is_liked,
        likes_count: Math.max(0, item.likes_count + (item.is_liked ? -1 : 1)),
      } : item))
    } catch (error) {
      console.error('点赞失败:', error)
    }
  }

  const toggleCollect = async (id: number) => {
    if (!requireAuth()) return
    const work = works.find(item => item.id === id)
    if (!work) return
    try {
      await (work.is_collected ? apiClient.works.uncollect(id) : apiClient.works.collect(id))
      setWorks(items => items.map(item => item.id === id ? {
        ...item,
        is_collected: !item.is_collected,
        collects_count: Math.max(0, item.collects_count + (item.is_collected ? -1 : 1)),
      } : item))
    } catch (error) {
      console.error('收藏失败:', error)
    }
  }

  const createTask = async (input: CreateTaskInput) => {
    if (!requireAuth()) return 'failed'
    try {
      if (input.type === 'image') {
        await apiClient.generation.createImage({ prompt: input.prompt, model_name: input.modelName, params: input.params, reference_image: input.referenceImage })
      } else {
        await apiClient.generation.createVideo({
          prompt: input.prompt,
          model_name: input.modelName,
          params: input.params,
          reference_images: input.referenceImages,
          first_frame_url: input.firstFrameUrl,
          last_frame_url: input.lastFrameUrl,
        })
      }
      await Promise.all([loadTasks(), loadUserProfile()])
      return 'success'
    } catch (error) {
      console.error('创建任务失败:', error)
      if (error instanceof ApiError && Number(error.code) === 5001) {
        window.dispatchEvent(new Event(CREDIT_PACKAGES_EVENT))
        return 'insufficient_credit'
      }
      return 'failed'
    }
  }

  const value = useMemo<Store>(() => ({
    user,
    authReady,
    balance: user?.balance ?? 0,
    works,
    tasks,
    models,
    ledger: [],
    authenticate,
    register,
    logout,
    requireAuth,
    toggleLike,
    toggleCollect,
    createTask,
    retryTask: () => undefined,
    recharge: () => undefined,
    checkIn: () => false,
    redeem: () => false,
    loadTasks,
	loadMoreTasks,
	hasMoreTasks: tasks.length < taskTotal,
	loadingMoreTasks,
	loadWorks,
	refreshProfile: loadUserProfile,
  }), [authReady, loadMoreTasks, loadTasks, loadWorks, loadingMoreTasks, models, taskTotal, tasks, user, works])

  return <Context.Provider value={value}>{children}</Context.Provider>
}

export const useApp = () => {
  const value = useContext(Context)
  if (!value) throw new Error('AppProvider missing')
  return value
}
