import { DownloadOutlined, EditOutlined, PlayCircleOutlined, RedoOutlined, ShareAltOutlined, ToolOutlined } from '@ant-design/icons'
import { Modal, Tooltip, message } from 'antd'
import { useEffect, useState } from 'react'
import { GeneratePanel } from '../features/creation'
import { useApp } from '../store/AppStore'
import type { Task } from '../types'
import { apiClient } from '../utils/api'

const actions=[
 [DownloadOutlined,'下载'],[RedoOutlined,'重新生成'],[EditOutlined,'以此编辑'],[ShareAltOutlined,'分享'],[ToolOutlined,'工具'],
] as const

const videoThumbnail = (task: Task) => task.thumbnail_url && task.thumbnail_url !== task.result_url ? task.thumbnail_url : ''

function GenerationTurn({task,onRegenerate,onPreview}:{task:Task;onRegenerate:(task:Task)=>void;onPreview:(task:Task)=>void}){
 const [msg,ctx]=message.useMessage()
 const action=async(title:string)=>{if(title==='下载'){if(!task.result_url)return msg.warning('任务结果尚不可下载');try{await apiClient.tasks.download(task.id);msg.success('开始下载')}catch(error){msg.error(error instanceof Error?error.message:'下载失败')}return}if(title==='重新生成'){onRegenerate(task);msg.success('内容已回填，可修改后重新生成');return}msg.success(`${title}操作暂未开放`)}
 const thumbnail=videoThumbnail(task),imageSource=task.thumbnail_url||task.result_url||'',canPreviewVideo=task.type==='video'&&task.status==='success'&&Boolean(task.result_url),canPreviewImage=task.type==='image'&&task.status==='success'&&Boolean(task.result_url)
 return <article className="generation-turn">{ctx}
  <div className="generation-prompt">{task.prompt}</div>
  <div className="generation-answer">
   <div className="generation-avatar">≈</div>
    <div className="generation-result">
    <div className={`generation-media ${task.status}`}>
     {task.type==='video'?<button type="button" className="generation-video-trigger" disabled={!canPreviewVideo} aria-label="播放视频" title={canPreviewVideo?'播放视频':undefined} onClick={()=>canPreviewVideo&&onPreview(task)}>{thumbnail?<img src={thumbnail} alt={task.title}/>:<span className="generation-video-label">{canPreviewVideo?'视频':'等待视频'}</span>}{canPreviewVideo&&<span className="generation-video-play"><PlayCircleOutlined/></span>}</button>:<button type="button" className="generation-image-trigger" disabled={!canPreviewImage} aria-label="预览图片" title={canPreviewImage?'预览图片':undefined} onClick={()=>canPreviewImage&&onPreview(task)}>{imageSource?<img src={imageSource} alt={task.title}/>:<span className="generation-video-label">等待图片</span>}</button>}
     {task.status==='processing'&&<div className="generation-progress"><span style={{width:`${task.progress}%`}}/><b>{task.progress}%</b></div>}
     {task.status==='failed'&&<div className="generation-failed">生成失败</div>}
    </div>
    <div className="generation-actions">{actions.map(([Icon,title])=><Tooltip title={title} key={title}><button aria-label={title} disabled={title==='下载'&&(!task.result_url||task.status!=='success')} onClick={()=>void action(title)}><Icon/></button></Tooltip>)}</div>
    <div className="generation-cost">{task.status==='processing'?`生成中 ${task.progress}% · `:task.status==='failed'?'生成失败 · ':''}消耗 {task.cost}</div>
   </div>
  </div>
 </article>
}

export function CreatePage(){
 const {tasks,loadTasks}=useApp()
 const [refill,setRefill]=useState<{prompt:string;key:number}>()
 const [preview,setPreview]=useState<Task|null>(null)
 useEffect(()=>{void loadTasks()},[loadTasks])
 const today=tasks.filter(t=>new Date(t.created_at).toDateString()===new Date().toDateString())
 const earlier=tasks.filter(t=>!today.includes(t))
 const regenerate=(task:Task)=>setRefill({prompt:task.prompt,key:Date.now()})
 return <main className="generation-page">
  <div className="generation-scroll">
   {tasks.length===0?<div className="generation-empty"><div>≈</div><h1>还没有生成记录</h1><p>在下方输入提示词开始创作，或前往灵感页获取灵感</p></div>:<div className="generation-thread">
   {today.length>0&&<section className="generation-day"><h1>今天</h1>{today.map(t=><GenerationTurn task={t} onRegenerate={regenerate} onPreview={setPreview} key={t.id}/>)}</section>}
    {earlier.length>0&&<section className="generation-day"><h1>更早</h1>{earlier.map(t=><GenerationTurn task={t} onRegenerate={regenerate} onPreview={setPreview} key={t.id}/>)}</section>}
   </div>}
  </div>
  <div className="generation-composer"><div className="generation-notice">◉ 新一代图像模型已上线，中文文字与细节表现全面提升</div><GeneratePanel refill={refill}/></div>
  <Modal className="generation-video-modal" title={preview?.type==='video'?'生成视频':'生成图片'} open={Boolean(preview)} footer={null} onCancel={()=>setPreview(null)} destroyOnHidden>{preview?.type==='video'?<video src={preview.result_url} poster={videoThumbnail(preview)||undefined} controls preload="metadata">当前浏览器不支持视频播放。</video>:preview&&<img src={preview.result_url} alt={preview.title} className="generation-image-preview"/>}</Modal>
 </main>
}
