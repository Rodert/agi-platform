import { PictureOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { Input, Modal, Select, message } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import type { Task } from '../../types'
import { apiClient } from '../../utils/api'
import { LazyImage } from '../../components/LazyImage'
import './publish-work-modal.css'

const workCategories = ['人物', '风景', '产品', '插画', '创意', '其他']

interface PublishWorkModalProps {
  open: boolean
  tasks: Task[]
  defaultTaskId?: number
  lockedTask?: Task
  modalTitle?: string
  onClose: () => void
  onPublished?: () => void | Promise<void>
}

const getThumbnail = (task: Task) => task.thumbnail_url || (task.type === 'image' ? task.result_url : '') || ''

export function PublishWorkModal({ open, tasks, defaultTaskId, lockedTask, modalTitle = '发布作品', onClose, onPublished }: PublishWorkModalProps) {
  const [selectedTaskId, setSelectedTaskId] = useState<number>()
  const [title, setTitle] = useState('')
  const [category, setCategory] = useState<string>()
  const [publishing, setPublishing] = useState(false)
  const [messageApi, contextHolder] = message.useMessage()
  const completedTasks = useMemo(() => tasks.filter(task => task.status === 'success' && task.result_url), [tasks])
  const availableTasks = useMemo(() => lockedTask ? [lockedTask] : completedTasks, [completedTasks, lockedTask])
  const selectedTask = availableTasks.find(task => task.id === selectedTaskId)

  useEffect(() => {
    if (!open) return
    const initialTask = lockedTask || availableTasks.find(task => task.id === defaultTaskId) || availableTasks[0]
    setSelectedTaskId(initialTask?.id)
    setTitle(initialTask?.title || initialTask?.prompt || '')
    setCategory(undefined)
  }, [availableTasks, defaultTaskId, lockedTask, open])

  const selectTask = (task: Task) => {
    setSelectedTaskId(task.id)
    setTitle(task.title || task.prompt)
  }

  const publish = async () => {
    if (!selectedTask) return messageApi.warning('请选择要发布的作品')
    if (!title.trim()) return messageApi.warning('请输入作品标题')
    setPublishing(true)
    try {
      await apiClient.works.publish({ task_id: selectedTask.id, title: title.trim(), category })
      messageApi.success('作品已提交，等待管理员审核')
      onClose()
      await onPublished?.()
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '发布失败')
    } finally {
      setPublishing(false)
    }
  }

  return <Modal className="publish-work-modal" title={modalTitle} open={open} okText="提交审核" cancelText="取消" okButtonProps={{ disabled: !selectedTask }} confirmLoading={publishing} onOk={() => void publish()} onCancel={onClose} destroyOnHidden>
    {contextHolder}
    {availableTasks.length === 0 ? <div className="publish-work-empty">暂无可发布的作品。完成图片或视频生成后，可在这里提交审核。</div> : <div className="publish-work-form">
      {lockedTask ? <div>
        <div className="publish-work-label">当前作品</div>
        <div className="publish-work-task is-selected publish-work-task-locked">
          <span className="publish-work-preview">{getThumbnail(lockedTask) ? <LazyImage src={getThumbnail(lockedTask)} alt="" /> : lockedTask.type === 'video' ? <PlayCircleOutlined /> : <PictureOutlined />}</span>
          <span className="publish-work-task-copy"><b>{lockedTask.title || (lockedTask.type === 'image' ? '图片作品' : '视频作品')}</b><small>{lockedTask.type === 'image' ? '图片' : '视频'} · {lockedTask.prompt}</small></span>
        </div>
      </div> : <div>
        <div className="publish-work-label">选择作品</div>
        <div className="publish-work-task-list">
          {availableTasks.map(task => {
            const thumbnail = getThumbnail(task)
            const selected = task.id === selectedTaskId
            return <button type="button" className={`publish-work-task${selected ? ' is-selected' : ''}`} key={task.id} onClick={() => selectTask(task)}>
              <span className="publish-work-preview">{thumbnail ? <LazyImage src={thumbnail} alt="" /> : task.type === 'video' ? <PlayCircleOutlined /> : <PictureOutlined />}</span>
              <span className="publish-work-task-copy"><b>{task.title || (task.type === 'image' ? '图片作品' : '视频作品')}</b><small>{task.type === 'image' ? '图片' : '视频'} · {task.prompt}</small></span>
              <span className="publish-work-radio" aria-hidden="true" />
            </button>
          })}
        </div>
      </div>}
      <div className="publish-work-fields">
        <div><div className="publish-work-label">作品标题</div><Input value={title} maxLength={255} placeholder="为作品填写标题" onChange={event => setTitle(event.target.value)} /></div>
        <div><div className="publish-work-label">分类</div><Select value={category} allowClear placeholder="选择分类（可选）" options={workCategories.map(value => ({ label: value, value }))} onChange={setCategory} /></div>
      </div>
    </div>}
  </Modal>
}
