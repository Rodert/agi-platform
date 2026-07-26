import { CopyOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { Button, Modal, QRCode, message } from 'antd'
import type { Task } from '../../types'
import { LazyImage } from '../../components/LazyImage'
import './share-work-modal.css'

interface ShareWorkModalProps {
  task: Task | null
  onClose: () => void
}

const getPoster = (task: Task) => task.thumbnail_url || (task.type === 'image' ? task.result_url : '') || ''

export function ShareWorkModal({ task, onClose }: ShareWorkModalProps) {
  const [messageApi, contextHolder] = message.useMessage()
  if (!task) return null
  // QR codes have a finite payload capacity. Use a stable short URL until the API provides a share token.
  const shareUrl = window.location.origin
  const copyUrl = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl)
      messageApi.success('分享链接已复制')
    } catch {
      messageApi.error('复制失败，请手动复制链接')
    }
  }
  const poster = getPoster(task)

  return <Modal className="share-work-modal" title="分享作品" open footer={<Button icon={<CopyOutlined />} onClick={() => void copyUrl()}>复制链接</Button>} onCancel={onClose} destroyOnHidden>
    {contextHolder}
    <div className="share-poster">
      <div className="share-poster-brand">潮汐 AI <span>AI CREATION</span></div>
      <div className="share-poster-media">{poster ? <LazyImage src={poster} alt={task.title || '分享作品'} /> : task.type === 'video' ? <span><PlayCircleOutlined /> 视频作品</span> : <span>图片作品</span>}</div>
      <div className="share-poster-caption"><b>{task.title || '我的 AI 作品'}</b><span>{task.type === 'video' ? 'AI 视频创作' : 'AI 图像创作'}</span></div>
      <div className="share-poster-footer"><QRCode value={shareUrl} size={68} bordered={false} color="#102326" bgColor="#edf5f3" /><div><b>扫码打开潮汐 AI</b><span>{window.location.host}</span></div></div>
    </div>
  </Modal>
}
