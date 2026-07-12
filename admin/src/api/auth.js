import request from '@/utils/request'

// 管理员登录
export const login = (username, password) => {
  return request.post('/auth/login', { username, password })
}
