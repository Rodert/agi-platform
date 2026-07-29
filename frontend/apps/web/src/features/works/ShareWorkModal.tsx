import { CopyOutlined, DownloadOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { Button, Modal, QRCode, message } from 'antd'
import { useRef, useState } from 'react'
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
  const [saving, setSaving] = useState(false)
  const posterRef = useRef<HTMLDivElement>(null)
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
  const savePoster = async () => {
    if (!posterRef.current) return
    setSaving(true)
    try {
      const width = 1080, padding = 56, mediaHeight = 760, scale = 2
      const canvas = document.createElement('canvas')
      canvas.width = width * scale
      canvas.height = 1240 * scale
      const context = canvas.getContext('2d')
      if (!context) throw new Error('浏览器不支持海报保存')
      context.scale(scale, scale)
      context.fillStyle = '#102326'
      context.fillRect(0, 0, width, 1240)
      context.fillStyle = '#a9e7df'
      context.font = '700 32px sans-serif'
      context.fillText('潮汐 AI', padding, 68)
      context.fillStyle = '#78aaa6'
      context.font = '600 20px sans-serif'
      context.textAlign = 'right'
      context.fillText('AI CREATION', width - padding, 68)
      context.textAlign = 'left'
      context.fillStyle = '#0b1718'
      context.fillRect(padding, 104, width - padding * 2, mediaHeight)
      if (poster) {
        const image = new Image()
        image.crossOrigin = 'anonymous'
        image.src = poster
        await image.decode()
        const ratio = Math.max((width - padding * 2) / image.naturalWidth, mediaHeight / image.naturalHeight)
        const imageWidth = image.naturalWidth * ratio, imageHeight = image.naturalHeight * ratio
        context.drawImage(image, padding + ((width - padding * 2) - imageWidth) / 2, 104 + (mediaHeight - imageHeight) / 2, imageWidth, imageHeight)
      } else {
        context.fillStyle = '#a9e7df'
        context.font = '500 32px sans-serif'
        context.textAlign = 'center'
        context.fillText(task.type === 'video' ? '视频作品' : '图片作品', width / 2, 104 + mediaHeight / 2)
        context.textAlign = 'left'
      }
      context.fillStyle = '#ffffff'
      context.font = '700 34px sans-serif'
      context.fillText(task.title || '我的 AI 作品', padding, 930)
      context.fillStyle = '#9fc4c0'
      context.font = '500 24px sans-serif'
      context.fillText(task.type === 'video' ? 'AI 视频创作' : 'AI 图像创作', padding, 974)
      context.fillStyle = '#edf5f3'
      context.fillRect(0, 1024, width, 216)
      const qrCanvas = posterRef.current.querySelector('canvas')
      if (qrCanvas) context.drawImage(qrCanvas, padding, 1098, 100, 100)
      context.fillStyle = '#102326'
      context.font = '700 26px sans-serif'
      context.fillText('扫码打开潮汐 AI', 184, 1145)
      context.fillStyle = '#547370'
      context.font = '500 21px sans-serif'
      context.fillText(window.location.host, 184, 1182)
      const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, 'image/png'))
      if (!blob) throw new Error('海报生成失败')
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `潮汐AI-作品-${task.id}.png`
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(link.href)
      messageApi.success('海报已开始保存')
    } catch {
      messageApi.error('海报保存失败，请检查作品资源的跨域配置')
    } finally {
      setSaving(false)
    }
  }

  return <Modal className="share-work-modal" title="分享作品" open footer={<><Button icon={<CopyOutlined />} onClick={() => void copyUrl()}>复制链接</Button><Button type="primary" icon={<DownloadOutlined />} loading={saving} onClick={() => void savePoster()}>保存海报</Button></>} onCancel={onClose} destroyOnHidden>
    {contextHolder}
    <div className="share-poster" ref={posterRef}>
      <div className="share-poster-brand">潮汐 AI <span>AI CREATION</span></div>
      <div className="share-poster-media">{poster ? <LazyImage src={poster} alt={task.title || '分享作品'} /> : task.type === 'video' ? <span><PlayCircleOutlined /> 视频作品</span> : <span>图片作品</span>}</div>
      <div className="share-poster-caption"><b>{task.title || '我的 AI 作品'}</b><span>{task.type === 'video' ? 'AI 视频创作' : 'AI 图像创作'}</span></div>
      <div className="share-poster-footer"><QRCode value={shareUrl} size={68} bordered={false} color="#102326" bgColor="#edf5f3" /><div><b>扫码打开潮汐 AI</b><span>{window.location.host}</span></div></div>
    </div>
  </Modal>
}
