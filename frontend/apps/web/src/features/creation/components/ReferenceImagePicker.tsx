import { CloseOutlined, FileImageOutlined, PlusOutlined } from '@ant-design/icons'
import { message } from 'antd'
import { useRef } from 'react'

const MAX_FILE_SIZE = 5 * 1024 * 1024
const ACCEPTED_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp'])

export function ReferenceImagePicker({ value, onChange, max = 1 }: { value: string[]; onChange: (value: string[]) => void; max?: number }) {
  const inputRef = useRef<HTMLInputElement>(null)
  const selectFiles = async (files: FileList | null) => {
    if (!files) return
    const selected = Array.from(files).slice(0, max - value.length)
    const valid = selected.filter(file => {
      if (!ACCEPTED_TYPES.has(file.type)) { message.warning('参考图仅支持 JPG、PNG 或 WebP'); return false }
      if (file.size > MAX_FILE_SIZE) { message.warning('单张参考图不能超过 5MB'); return false }
      return true
    })
    const images = await Promise.all(valid.map(file => new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result))
      reader.onerror = () => reject(reader.error)
      reader.readAsDataURL(file)
    })))
    if (images.length) onChange([...value, ...images])
    if (inputRef.current) inputRef.current.value = ''
  }
  return <div className="reference-image-picker">
    {value.map((url, index) => <div className="reference-image-thumb" key={url}><img src={url} alt={`参考图 ${index + 1}`} /><button type="button" aria-label="移除参考图" onClick={() => onChange(value.filter((_, itemIndex) => itemIndex !== index))}><CloseOutlined /></button></div>)}
    {value.length < max && <button type="button" className="reference-image-add" onClick={() => inputRef.current?.click()}><FileImageOutlined /><span>{value.length ? <PlusOutlined /> : '参考图'}</span></button>}
    <input ref={inputRef} className="hidden" type="file" accept="image/jpeg,image/png,image/webp" multiple={max > 1} onChange={event => void selectFiles(event.target.files)} />
  </div>
}
