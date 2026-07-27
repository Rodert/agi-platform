import { message } from 'antd'
import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { modelParamDefaults } from '../creation/components/ModelParams'
import { useApp } from '../../store/AppStore'
import type { Work } from '../../types'

export function useCreateFromWork() {
  const { createTask, models, requireAuth } = useApp()
  const navigate = useNavigate()

  return useCallback(async (work: Work, onCreated?: () => void) => {
    if (!requireAuth()) return false

    const model = models.find(item => item.type === work.type)
    if (!model) {
      message.error(`暂无可用${work.type === 'video' ? '视频' : '图片'}模型`)
      return false
    }

    const params = modelParamDefaults(model)
    const ratioOptions = model.params_config?.ratio?.options ?? []
    if (work.ratio && ratioOptions.some(option => option.value === work.ratio)) {
      params.ratio = work.ratio
    }

    const result = await createTask({
      prompt: work.prompt,
      type: work.type,
      modelName: model.name,
      params,
    })
    if (result !== 'success') {
      if (result === 'failed') message.error('任务提交失败，请稍后重试')
      return false
    }

    onCreated?.()
    navigate('/create')
    return true
  }, [createTask, models, navigate, requireAuth])
}
