import { DownloadOutlined, RedoOutlined, StopOutlined } from '@ant-design/icons'
import { Button, Progress, Segmented, Tag, message } from 'antd'
import { useState } from 'react'
import { LazyImage } from '../components/LazyImage'
import { useApp } from '../store/AppStore'
import { apiClient } from '../utils/api'

const statusLabel={queued:'排队中',processing:'生成中',success:'已完成',failed:'失败'}

export function TasksPage(){
 const {tasks,retryTask}=useApp()
 const [status,setStatus]=useState('全部')
 const [msg,ctx]=message.useMessage()
 const list=tasks.filter(t=>status==='全部'||statusLabel[t.status]===status)
 const download=async(id:number)=>{try{await apiClient.tasks.download(id);msg.success('开始下载')}catch(error){msg.error(error instanceof Error?error.message:'下载失败')}}
 return <main className="page">{ctx}<h1 className="page-title">创作任务</h1><p className="page-subtitle">生成结果仅保留 7 天，请在完成后尽快下载保存</p><div className="toolbar"><Segmented options={['全部','排队中','生成中','已完成','失败']} value={status} onChange={v=>setStatus(String(v))}/></div><div className="grid gap-3">{list.map(t=><article key={t.id} className="task-row surface flex items-center gap-4 p-4 max-[650px]:items-start">{(t.thumbnail_url||t.result_url)?<LazyImage src={t.thumbnail_url||t.result_url} alt="" className="h-20 w-20 rounded-[10px] object-cover"/>:<div className="theme-raised theme-muted grid h-20 w-20 place-items-center rounded-[10px] text-xs">暂无预览</div>}<div className="min-w-0 flex-1"><div className="flex justify-between gap-3"><b className="truncate">{t.title||t.prompt}</b><Tag color={t.status==='success'?'success':t.status==='failed'?'error':'warning'}>{statusLabel[t.status]}</Tag></div><div className="theme-muted my-2 text-xs">{t.type==='image'?'图片生成':'视频生成'} · {new Date(t.created_at).toLocaleString('zh-CN')} · 消耗 {t.cost}</div>{t.status==='failed'&&<div className="mt-2 text-xs text-red-400">{t.error_msg||'生成失败，请稍后重试'}</div>}{t.status==='processing'&&<Progress percent={t.progress} size="small" strokeColor="#54d6cf"/>}</div><div className="flex gap-2 max-[650px]:flex-col">{t.status==='failed'&&<Button icon={<RedoOutlined/>} onClick={()=>retryTask(t.id)} disabled>暂不支持重试</Button>}{t.status==='processing'&&<Button icon={<StopOutlined/>} disabled>取消</Button>}{t.status==='success'&&t.result_url&&<Button icon={<DownloadOutlined/>} onClick={()=>void download(t.id)}>下载保存</Button>}</div></article>)}</div>{!list.length&&<div className="theme-muted surface py-16 text-center">暂无任务</div>}</main>
}
