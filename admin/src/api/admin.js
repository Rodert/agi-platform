import request from '@/utils/request'

// 获取统计数据
export const getStats = () => {
  return request.get('/stats')
}

// 查询管理端操作审计日志。
export const getLogs = (params = {}) => {
  return request.get('/logs', { params })
}

export const getReport = (startDate, endDate) => {
  return request.get('/reports', { params: { start_date: startDate, end_date: endDate } })
}

export const getDatabaseTables = () => request.get('/database/tables')
export const getDatabaseTable = (table, params = {}) => request.get(`/database/tables/${encodeURIComponent(table)}`, { params })

// 获取待审核作品列表
export const getPendingWorks = (page = 1, pageSize = 20) => {
  return request.get('/works/pending', { params: { page, page_size: pageSize } })
}

// 获取全部作品，可按生命周期状态筛选。
export const getWorks = (params = {}) => {
  return request.get('/works', { params })
}

// 审核作品
export const auditWork = (id, status, reason = '') => {
  return request.post(`/works/${id}/audit`, { status, reason })
}

// 下架或重新上架已审核作品，不会删除长期资源。
export const updateWorkStatus = (id, status, reason = '') => {
  return request.post(`/works/${id}/status`, { status, reason })
}

export const getAdmins = () => request.get('/admins')
export const createAdmin = (data) => request.post('/admins', data)
export const updateAdmin = (id, data) => request.put(`/admins/${id}`, data)
export const getSystemUpdateStatus = () => request.get('/system/update')
export const triggerSystemUpdate = () => request.post('/system/update')
