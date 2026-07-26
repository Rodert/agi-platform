import { AppstoreAddOutlined, DownloadOutlined, FileImageOutlined, FontColorsOutlined, NodeIndexOutlined, UploadOutlined } from '@ant-design/icons'
import { Button, Empty, Modal, Select, Slider, Upload, message } from 'antd'
import type { UploadProps } from 'antd'
import { useEffect, useMemo, useState } from 'react'

type ToolKind = 'prompt' | 'svg' | 'convert' | 'more'
type ImageFormat = 'image/jpeg' | 'image/png' | 'image/webp'

const tools: { kind: ToolKind; title: string; description: string; Icon: typeof FontColorsOutlined }[] = [
  { kind: 'prompt', title: '图片反推提示词', description: '上传图片，AI 自动反推生成提示词', Icon: FontColorsOutlined },
  { kind: 'svg', title: '图片转 SVG', description: '高质量矢量化转换，即将推出', Icon: NodeIndexOutlined },
  { kind: 'convert', title: '图片格式转换', description: '转换并压缩 PNG、JPG、WebP 图片', Icon: FileImageOutlined },
  { kind: 'more', title: '更多工具即将推出', description: '持续扩展实用的创作辅助能力', Icon: AppstoreAddOutlined },
]

const formatNames: Record<ImageFormat, string> = { 'image/jpeg': 'JPG', 'image/png': 'PNG', 'image/webp': 'WebP' }
const acceptedTypes = new Set<ImageFormat>(['image/jpeg', 'image/png', 'image/webp'])

function fileBaseName(name: string) { return name.replace(/\.[^.]+$/, '') || 'image' }
function download(blob: Blob, name: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = name
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
function loadImage(source: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片读取失败'))
    image.src = source
  })
}
export function ToolsPage() {
  const [active, setActive] = useState<ToolKind | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [source, setSource] = useState('')
  const [format, setFormat] = useState<ImageFormat>('image/webp')
  const [quality, setQuality] = useState(85)
  const [processing, setProcessing] = useState(false)
  const [msg, context] = message.useMessage()
  const activeTool = useMemo(() => tools.find(tool => tool.kind === active), [active])

  useEffect(() => () => { if (source) URL.revokeObjectURL(source) }, [source])
  const close = () => { setActive(null); setFile(null); setSource('') }
  const selectFile: UploadProps['beforeUpload'] = candidate => {
    if (!acceptedTypes.has(candidate.type as ImageFormat)) { msg.warning('请选择 JPG、PNG 或 WebP 图片'); return Upload.LIST_IGNORE }
    if (candidate.size > 15 * 1024 * 1024) { msg.warning('图片不能超过 15MB'); return Upload.LIST_IGNORE }
    if (source) URL.revokeObjectURL(source)
    setFile(candidate)
    setSource(URL.createObjectURL(candidate))
    return false
  }
  const processImage = async () => {
    if (!file || !source || !active) return
    setProcessing(true)
    try {
      const image = await loadImage(source)
      if (active === 'convert') {
        const canvas = document.createElement('canvas')
        canvas.width = image.naturalWidth
        canvas.height = image.naturalHeight
        const context = canvas.getContext('2d')
        if (!context) throw new Error('浏览器不支持图片处理')
        if (format === 'image/jpeg') { context.fillStyle = '#ffffff'; context.fillRect(0, 0, canvas.width, canvas.height) }
        context.drawImage(image, 0, 0)
        const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, format, quality / 100))
        if (!blob) throw new Error('图片转换失败')
        const extension = format === 'image/jpeg' ? 'jpg' : format.slice(6)
        download(blob, `${fileBaseName(file.name)}.${extension}`)
        msg.success(`已导出 ${formatNames[format]} 图片`)
      }
    } catch (error) { msg.error(error instanceof Error ? error.message : '处理失败') } finally { setProcessing(false) }
  }

  return <main className="page">{context}
    <h1 className="page-title">工具箱</h1><p className="page-subtitle">实用的创作辅助工具，图片全程在本地浏览器处理</p>
    <div className="mt-7 grid grid-cols-3 gap-4 max-[850px]:grid-cols-2 max-[520px]:grid-cols-1">{tools.map(tool => {
      const Icon = tool.Icon
      return <button key={tool.kind} onClick={() => setActive(tool.kind)} className="surface tool-card p-5 text-left transition"><span className="tool-icon grid h-11 w-11 place-items-center rounded-lg text-xl"><Icon /></span><span><b className="block text-base">{tool.title}</b><small className="theme-muted mt-1 block">{tool.description}</small></span></button>
    })}</div>
    <Modal open={!!active} title={activeTool?.title} onCancel={close} footer={null} destroyOnHidden>
      {active === 'prompt' || active === 'svg' || active === 'more' ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={active === 'svg' ? '图片转 SVG 即将推出' : active === 'more' ? '更多工具正在开发中' : '图片反推提示词暂未开放'} className="py-8" /> : <>
        <p className="theme-muted mb-4">{activeTool?.description}</p>
        <Upload.Dragger accept="image/jpeg,image/png,image/webp" beforeUpload={selectFile} showUploadList={false} maxCount={1} disabled={processing}>
          {source ? <img src={source} alt="待处理图片" className="mx-auto max-h-52 max-w-full rounded-md object-contain" /> : <><UploadOutlined className="theme-accent text-2xl" /><p className="mt-2">点击或拖入图片</p><p className="theme-muted text-xs">支持 JPG、PNG、WebP，最大 15MB</p></>}
        </Upload.Dragger>
        {file && <p className="theme-muted mt-3 truncate text-xs">{file.name} · {(file.size / 1024 / 1024).toFixed(2)} MB</p>}
        {active === 'convert' && <div className="mt-5 grid grid-cols-[1fr_1.5fr] gap-4 max-[460px]:grid-cols-1"><label className="theme-muted"><span className="mb-2 block text-xs">导出格式</span><Select value={format} onChange={setFormat} className="w-full" options={Object.entries(formatNames).map(([value, label]) => ({ value, label }))} /></label><label className="theme-muted"><span className="mb-2 flex justify-between text-xs">压缩质量 <b className="theme-accent">{quality}</b></span><Slider value={quality} onChange={setQuality} min={10} max={100} disabled={format === 'image/png'} /></label></div>}
        <Button block type="primary" icon={<DownloadOutlined />} className="!mt-6" onClick={() => void processImage()} disabled={!file} loading={processing}>转换并下载</Button>
      </>}
    </Modal>
  </main>
}
