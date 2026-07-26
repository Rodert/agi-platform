import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login as loginApi } from '@/api/auth'

const readAdminInfo = () => {
  const value = localStorage.getItem('admin_info')
  if (!value || value === 'undefined') {
    localStorage.removeItem('admin_info')
    return null
  }
  try {
    return JSON.parse(value)
  } catch {
    localStorage.removeItem('admin_info')
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const storedToken = localStorage.getItem('admin_token')
  const token = ref(storedToken && storedToken !== 'undefined' ? storedToken : '')
  const adminInfo = ref(readAdminInfo())

  if (!token.value) {
    localStorage.removeItem('admin_token')
  }

  const login = async (username, password) => {
    const res = await loginApi(username, password)
    token.value = res.token
    adminInfo.value = res.admin

    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_info', JSON.stringify(res.admin))
  }

  const logout = () => {
    token.value = ''
    adminInfo.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
  }

  const updateAdminInfo = (profile) => {
    adminInfo.value = { ...adminInfo.value, ...profile }
    localStorage.setItem('admin_info', JSON.stringify(adminInfo.value))
  }

  return {
    token,
    adminInfo,
    login,
    logout,
    updateAdminInfo
  }
})
