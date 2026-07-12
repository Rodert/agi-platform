import { DeleteOutlined, DownloadOutlined, HeartFilled, HeartOutlined, ShareAltOutlined, ToolOutlined } from '@ant-design/icons'
import { Empty, Tooltip, message } from 'antd'
import { useMemo, useState } from 'react'
import { useApp } from '../../store/AppStore'
import type { Work } from '../../types'

const mediaTabs=['全部','图片','视频'] as const
const relationTabs=['所有','我的收藏','我的分享','我的喜欢'] as const

function AssetCard({work,onToggle,onDelete}:{work:Work;onToggle:()=>void;onDelete:()=>void}){
 const [msg,ctx]=message.useMessage()
 const simulate=(name:string)=>msg.success(`${name}操作已模拟`)
 return <article className="asset-card">{ctx}<div className="asset-media">
  <img src={work.image_url||work.video_url} alt={work.title}/>{work.type==='video'&&<span className="asset-video">▶ 视频</span>}
 </div><p>{work.title}</p><div className="asset-actions">
  <Tooltip title={work.is_collected?'取消收藏':'收藏'}><button aria-label={work.is_collected?'取消收藏':'收藏'} onClick={onToggle}>{work.is_collected?<HeartFilled className="text-[#ef7893]"/>:<HeartOutlined/>}</button></Tooltip>
  <Tooltip title="下载"><button aria-label="下载" onClick={()=>simulate('下载')}><DownloadOutlined/></button></Tooltip>
  <Tooltip title="分享"><button aria-label="分享" onClick={()=>simulate('分享')}><ShareAltOutlined/></button></Tooltip>
  <Tooltip title="删除"><button aria-label="删除" onClick={onDelete}><DeleteOutlined/></button></Tooltip>
  <Tooltip title="工具"><button aria-label="工具" onClick={()=>simulate('工具')}><ToolOutlined/></button></Tooltip>
 </div></article>
}

export function AssetsPage(){
 const {works,toggleCollect}=useApp()
 const [media,setMedia]=useState<(typeof mediaTabs)[number]>('全部')
 const [relation,setRelation]=useState<(typeof relationTabs)[number]>('所有')
 const [deleted,setDeleted]=useState<number[]>([])
 const [msg,ctx]=message.useMessage()
 const list=useMemo(()=>works.filter(work=>{
  if(deleted.includes(work.id))return false
  if(media!=='全部'&&work.type!==(media==='图片'?'image':'video'))return false
  if(relation==='我的收藏'&&!work.is_collected)return false
  if(relation==='我的喜欢'&&!work.is_liked)return false
  if(relation==='我的分享')return false
  return true
 }),[works,media,relation,deleted])
 const remove=(id:number)=>{setDeleted(v=>[...v,id]);msg.info('删除接口尚未开放')}
 return <main className="assets-page">{ctx}
  <div className="asset-filters">
   <div className="asset-media-tabs">{mediaTabs.map(tab=><button className={media===tab?'active':''} onClick={()=>setMedia(tab)} key={tab}>{tab}</button>)}</div>
   <div className="asset-relation-tabs">{relationTabs.map(tab=><button className={relation===tab?'active':''} onClick={()=>setRelation(tab)} key={tab}>{tab}</button>)}</div>
  </div>
  {list.length?<section className="asset-day"><h1>今天</h1><div className="asset-grid">{list.map(work=><AssetCard key={work.id} work={work} onToggle={()=>toggleCollect(work.id)} onDelete={()=>remove(work.id)}/>)}</div></section>:<div className="asset-empty"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无相关资产"/></div>}
 </main>
}
