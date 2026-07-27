import { Button, Input, Select, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useApp } from '../store/AppStore'
import { ModelParams, modelParamCost, modelParamDefaults } from '../features/creation/components/ModelParams'

export function VideoPage() {
  const [prompt, setPrompt] = useState('')
  const [model, setModel] = useState('')
  const [modelParams, setModelParams] = useState<Record<string, unknown>>({})
  const [msg, context] = message.useMessage()
  const navigate = useNavigate()
  const { createTask, models } = useApp()
  const videoModels = models.filter(item => item.type === 'video')
  const current = videoModels.find(item => item.name === model) || videoModels[0]
  useEffect(() => { setModelParams(modelParamDefaults(current)) }, [current])

  const submit = async () => {
    if (!prompt.trim()) return msg.warning('请输入视频描述')
    if (!current) return msg.error('暂无可用视频模型')
    const result = await createTask({
      prompt: prompt.trim(), type: 'video', modelName: current.name,
      params: modelParams,
    })
    if (result === 'success') { setPrompt(''); navigate('/create') } else if (result === 'failed') msg.error('任务提交失败')
  }

  return <main className="page">{context}<h1 className="page-title">视频生成</h1><p className="page-subtitle">根据文字描述生成视频</p><div className="mt-6 grid grid-cols-[1fr_340px] gap-5 max-[900px]:grid-cols-1"><section className="surface p-5"><Input.TextArea rows={8} value={prompt} onChange={event => setPrompt(event.target.value)} placeholder="描述镜头运动、主体动作与环境变化..."/><div className="mt-4 flex justify-end"><Button type="primary" disabled={!current} onClick={() => void submit()}>生成视频 · {current ? modelParamCost(current,modelParams) : '-'} 灵感值</Button></div></section><aside className="surface p-5"><h3 className="mt-0">视频参数</h3><label className="theme-muted mb-2 block">模型</label><Select value={current?.name} placeholder="暂无可用模型" onChange={setModel} options={videoModels.map(item => ({ label: item.display_name, value: item.name }))} className="mb-5 w-full"/><ModelParams model={current} values={modelParams} onChange={setModelParams}/></aside></div></main>
}
