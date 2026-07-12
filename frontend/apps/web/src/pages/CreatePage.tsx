import { DownloadOutlined, EditOutlined, RedoOutlined, ShareAltOutlined, ToolOutlined } from '@ant-design/icons'
import { Tooltip, message } from 'antd'
import { useState } from 'react'
import { GeneratePanel } from '../features/creation'
import { useApp } from '../store/AppStore'
import type { Task } from '../types'

const actions=[
 [DownloadOutlined,'下载'],[RedoOutlined,'重新生成'],[EditOutlined,'以此编辑'],[ShareAltOutlined,'分享'],[ToolOutlined,'工具'],
] as const

function GenerationTurn({task,onRegenerate}:{task:Task;onRegenerate:(task:Task)=>void}){
 const [msg,ctx]=message.useMessage()
 const action=(title:string)=>{if(title==='重新生成'){onRegenerate(task);msg.success('内容已回填，可修改后重新生成');return}msg.success(`${title}操作已模拟`)}
 return <article className="generation-turn">{ctx}
  <div className="generation-prompt">{task.prompt}</div>
  <div className="generation-answer">
   <div className="generation-avatar">≈</div>
   <div className="generation-result">
    <div className={`generation-media ${task.status}`}>
     {(task.thumbnail_url||task.result_url)?<img src={task.thumbnail_url||task.result_url} alt={task.title}/>:<div className="generation-failed">等待生成结果</div>}
     {task.status==='processing'&&<div className="generation-progress"><span style={{width:`${task.progress}%`}}/><b>{task.progress}%</b></div>}
     {task.status==='failed'&&<div className="generation-failed">生成失败</div>}
    </div>
    <div className="generation-actions">{actions.map(([Icon,title])=><Tooltip title={title} key={title}><button aria-label={title} onClick={()=>action(title)}><Icon/></button></Tooltip>)}</div>
    <div className="generation-cost">{task.status==='processing'?`生成中 ${task.progress}% · `:task.status==='failed'?'生成失败 · ':''}消耗 {task.cost}</div>
   </div>
  </div>
 </article>
}

export function CreatePage(){
 const {tasks}=useApp()
 const [refill,setRefill]=useState<{prompt:string;key:number}>()
 const today=tasks.filter(t=>new Date(t.created_at).toDateString()===new Date().toDateString())
 const earlier=tasks.filter(t=>!today.includes(t))
 const regenerate=(task:Task)=>setRefill({prompt:task.prompt,key:Date.now()})
 return <main className="generation-page">
  <div className="generation-scroll">
   {tasks.length===0?<div className="generation-empty"><div>≈</div><h1>还没有生成记录</h1><p>在下方输入提示词开始创作，或前往灵感页获取灵感</p></div>:<div className="generation-thread">
    {today.length>0&&<section className="generation-day"><h1>今天</h1>{today.map(t=><GenerationTurn task={t} onRegenerate={regenerate} key={t.id}/>)}</section>}
    {earlier.length>0&&<section className="generation-day"><h1>更早</h1>{earlier.map(t=><GenerationTurn task={t} onRegenerate={regenerate} key={t.id}/>)}</section>}
   </div>}
  </div>
  <div className="generation-composer"><div className="generation-notice">◉ 新一代图像模型已上线，中文文字与细节表现全面提升</div><GeneratePanel refill={refill}/></div>
 </main>
}
