import { CopyOutlined, DownloadOutlined, EditOutlined, FileImageOutlined, PlayCircleOutlined, RedoOutlined, ShareAltOutlined, ToolOutlined, UploadOutlined } from '@ant-design/icons'
import { Button, Modal, Slider, Tooltip, message } from 'antd'
import { useEffect, useState } from 'react'
import './CreatePage.css'
import { GeneratePanel, InlineVideoPanel, type CreationMode } from '../features/creation'
import { useApp } from '../store/AppStore'
import type { Task } from '../types'
import { apiClient } from '../utils/api'
import { PublishWorkModal, ShareWorkModal } from '../features/works'
import { LazyImage } from '../components/LazyImage'

const actions=[
 [DownloadOutlined,'下载'],[RedoOutlined,'重新生成'],[EditOutlined,'以此编辑'],[UploadOutlined,'发布作品'],[ShareAltOutlined,'分享'],[ToolOutlined,'工具'],
] as const

const videoThumbnail = (task: Task) => task.thumbnail_url && task.thumbnail_url !== task.result_url ? task.thumbnail_url : ''
const assetExpiry = (task: Task) => {
 const date=new Date(task.completed_at||task.created_at)
 date.setDate(date.getDate()+7)
 return date.toLocaleDateString('zh-CN',{month:'long',day:'numeric'})
}

function GenerationTurn({task,onRegenerate,onPreview,onPublish,onShare,onTools}:{task:Task;onRegenerate:(task:Task)=>void;onPreview:(task:Task)=>void;onPublish:(task:Task)=>void;onShare:(task:Task)=>void;onTools:(task:Task)=>void}){
 const [msg,ctx]=message.useMessage()
 const action=async(title:string)=>{if(title==='下载'){if(!task.result_url)return msg.warning('任务结果尚不可下载');try{await apiClient.tasks.download(task.id);msg.success('开始下载')}catch(error){msg.error(error instanceof Error?error.message:'下载失败')}return}if(title==='重新生成'){onRegenerate(task);msg.success('内容已回填，可修改后重新生成');return}if(title==='发布作品'){onPublish(task);return}if(title==='分享'){onShare(task);return}if(title==='工具'){if(task.type!=='image'||task.status!=='success'||!task.result_url)return msg.warning('图片生成完成后可使用工具');onTools(task);return}msg.success(`${title}操作暂未开放`)}
 const copyPrompt=async()=>{try{if(navigator.clipboard?.writeText){await navigator.clipboard.writeText(task.prompt)}else{const input=document.createElement('textarea');input.value=task.prompt;input.style.position='fixed';input.style.opacity='0';document.body.appendChild(input);input.select();const copied=document.execCommand('copy');document.body.removeChild(input);if(!copied)throw new Error('复制失败')}msg.success('提示词已复制')}catch{msg.error('复制失败，请手动复制')}}
 const thumbnail=videoThumbnail(task),imageSource=task.thumbnail_url||task.result_url||'',canPreviewVideo=task.type==='video'&&task.status==='success'&&Boolean(task.result_url),canPreviewImage=task.type==='image'&&task.status==='success'&&Boolean(task.result_url)
 return <article className="generation-turn">{ctx}
  <div className="generation-prompt">{task.prompt}<Tooltip title="复制提示词"><button type="button" className="generation-prompt-copy" aria-label="复制提示词" onClick={()=>void copyPrompt()}><CopyOutlined/></button></Tooltip></div>
  <div className="generation-answer">
   <div className="generation-avatar">≈</div>
    <div className="generation-result">
    <div className={`generation-media ${task.status}`}>
     {task.type==='video'?<button type="button" className="generation-video-trigger" disabled={!canPreviewVideo} aria-label="播放视频" title={canPreviewVideo?'播放视频':undefined} onClick={()=>canPreviewVideo&&onPreview(task)}>{thumbnail?<LazyImage rootSelector=".generation-scroll" src={thumbnail} alt={task.title}/>:<span className="generation-video-label">{canPreviewVideo?'视频':'等待视频'}</span>}{canPreviewVideo&&<span className="generation-video-play"><PlayCircleOutlined/></span>}</button>:<button type="button" className="generation-image-trigger" disabled={!canPreviewImage} aria-label="预览图片" title={canPreviewImage?'预览图片':undefined} onClick={()=>canPreviewImage&&onPreview(task)}>{imageSource?<LazyImage rootSelector=".generation-scroll" src={imageSource} alt={task.title}/>:<span className="generation-video-label">等待图片</span>}</button>}
     {task.status==='processing'&&<div className="generation-progress"><span style={{width:`${task.progress}%`}}/><b>{task.progress}%</b></div>}
     {task.status==='failed'&&<div className="generation-failed">生成失败</div>}
    </div>
    <div className="generation-actions">{actions.map(([Icon,title])=><Tooltip title={title} key={title}><button aria-label={title} disabled={((title==='下载'||title==='发布作品'||title==='分享')&&(!task.result_url||task.status!=='success'))||(title==='工具'&&(task.type!=='image'||task.status!=='success'||!task.result_url))} onClick={()=>void action(title)}><Icon/></button></Tooltip>)}</div>
	{task.status==='failed'&&<div className="generation-error" title={task.error_msg}>{task.error_msg||'生成失败，请稍后重试'}</div>}
    {task.status==='success'&&task.result_url&&<div className="generation-retention-notice"><span className="generation-retention-icon"><DownloadOutlined/></span><span><b>请尽快保存到本地</b><small>该{task.type==='video'?'视频':'图片'}将于 {assetExpiry(task)} 自动清理</small></span><button type="button" onClick={()=>void action('下载')}>立即下载</button></div>}
    <div className="generation-cost">{task.status==='processing'?`生成中 ${task.progress}% · `:task.status==='failed'?'生成失败 · ':''}消耗 {task.cost}</div>
   </div>
  </div>
 </article>
}

export function CreatePage(){
 const {tasks,loadTasks}=useApp()
 const [mode,setMode]=useState<CreationMode>('image')
 const [refill,setRefill]=useState<{prompt:string;key:number}>()
 const [preview,setPreview]=useState<Task|null>(null)
 const [publishTask,setPublishTask]=useState<Task|null>(null)
 const [shareTask,setShareTask]=useState<Task|null>(null)
 const [toolTask,setToolTask]=useState<Task|null>(null)
 const [toolQuality,setToolQuality]=useState(75)
 const [processingTool,setProcessingTool]=useState(false)
 const [msg,ctx]=message.useMessage()
 useEffect(()=>{void loadTasks()},[loadTasks])
 const today=tasks.filter(t=>new Date(t.created_at).toDateString()===new Date().toDateString())
 const earlier=tasks.filter(t=>!today.includes(t))
 const regenerate=(task:Task)=>setRefill({prompt:task.prompt,key:Date.now()})
 const processImage=async(format:'image/jpeg'|'image/webp')=>{
  if(!toolTask?.result_url)return
  setProcessingTool(true)
  try{
   const response=await fetch(toolTask.result_url)
   if(!response.ok)throw new Error('无法读取原图')
   const source=await response.blob()
   const bitmap=await createImageBitmap(source)
   const canvas=document.createElement('canvas')
   canvas.width=bitmap.width
   canvas.height=bitmap.height
   const context=canvas.getContext('2d')
   if(!context)throw new Error('浏览器不支持图片处理')
   if(format==='image/jpeg'){context.fillStyle='#ffffff';context.fillRect(0,0,canvas.width,canvas.height)}
   context.drawImage(bitmap,0,0)
   bitmap.close()
   const output=await new Promise<Blob|null>(resolve=>canvas.toBlob(resolve,format,toolQuality/100))
   if(!output)throw new Error('图片处理失败')
   const link=document.createElement('a')
   link.href=URL.createObjectURL(output)
   link.download=`agi-${toolTask.id}-${format==='image/webp'?'webp':'compressed'}.${format==='image/webp'?'webp':'jpg'}`
   document.body.appendChild(link)
   link.click()
   link.remove()
   URL.revokeObjectURL(link.href)
   msg.success(format==='image/webp'?'WebP 图片已开始下载':'压缩图片已开始下载')
  }catch(error){msg.error(error instanceof Error?`${error.message}，请检查对象存储的跨域配置`:'图片处理失败')}
  finally{setProcessingTool(false)}
 }
 return <main className="generation-page">
  {ctx}
  <div className="generation-scroll">
   {tasks.length===0?<div className="generation-empty"><div>≈</div><h1>还没有生成记录</h1><p>在下方输入提示词开始创作，或前往灵感页获取灵感</p></div>:<div className="generation-thread">
   {today.length>0&&<section className="generation-day"><h1>今天</h1>{today.map(t=><GenerationTurn task={t} onRegenerate={regenerate} onPreview={setPreview} onPublish={setPublishTask} onShare={setShareTask} onTools={setToolTask} key={t.id}/>)}</section>}
    {earlier.length>0&&<section className="generation-day"><h1>更早</h1>{earlier.map(t=><GenerationTurn task={t} onRegenerate={regenerate} onPreview={setPreview} onPublish={setPublishTask} onShare={setShareTask} onTools={setToolTask} key={t.id}/>)}</section>}
   </div>}
  </div>
  <div className="generation-composer"><div className="generation-notice">{mode==='image'?'◉ 新一代图像模型已上线，中文文字与细节表现全面提升':'◉ 描述主体、动作与镜头，让 AI 生成动态画面'}</div>{mode==='image'?<GeneratePanel refill={refill} onModeChange={setMode}/>:<InlineVideoPanel onModeChange={setMode}/>}</div>
  <Modal className="generation-video-modal" title={preview?.type==='video'?'生成视频':'生成图片'} open={Boolean(preview)} footer={null} onCancel={()=>setPreview(null)} destroyOnHidden>{preview?.type==='video'?<video src={preview.result_url} poster={videoThumbnail(preview)||undefined} controls preload="metadata">当前浏览器不支持视频播放。</video>:preview&&<img src={preview.result_url} alt={preview.title} className="generation-image-preview" decoding="async"/>}</Modal>
  <Modal title="图片工具" open={Boolean(toolTask)} onCancel={()=>setToolTask(null)} footer={null} destroyOnHidden>
   <div className="image-tool-list">
    <div className="image-tool-row"><span className="image-tool-icon"><FileImageOutlined/></span><div><b>压缩图片</b><p>保持原始尺寸，降低文件体积并下载 JPG。</p></div><Button type="primary" loading={processingTool} onClick={()=>void processImage('image/jpeg')}>压缩并下载</Button></div>
    <div className="image-tool-quality"><span>压缩质量</span><Slider min={20} max={95} value={toolQuality} onChange={setToolQuality} tooltip={{formatter:value=>`${value ?? toolQuality}%`}}/></div>
    <div className="image-tool-row"><span className="image-tool-icon"><DownloadOutlined/></span><div><b>转为 WebP</b><p>以 WebP 格式导出，适合网页和内容发布。</p></div><Button loading={processingTool} onClick={()=>void processImage('image/webp')}>转换并下载</Button></div>
   </div>
  </Modal>
  <PublishWorkModal open={Boolean(publishTask)} tasks={tasks} lockedTask={publishTask || undefined} onClose={()=>setPublishTask(null)} onPublished={loadTasks}/>
  <ShareWorkModal task={shareTask} onClose={()=>setShareTask(null)}/>
 </main>
}
