import request from '@/utils/request'

// 获取统计数据
export const getStats = () => {
  return request.get('/stats')
}

// 获取待审核作品列表
export const getPendingWorks = (page = 1, pageSize = 20) => {
  return request.get('/works/pending', { params: { page, page_size: pageSize } })
}

// 审核作品
export const auditWork = (id, status, reason = '') => {
  return request.post(`/works/${id}/audit`, { status, reason })
}
