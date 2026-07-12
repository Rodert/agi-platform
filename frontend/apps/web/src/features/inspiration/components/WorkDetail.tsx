import { CloseOutlined, CopyOutlined, HeartFilled, HeartOutlined } from '@ant-design/icons'
import { message } from 'antd'
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useApp } from '../../../store/AppStore'
import type { Work } from '../../../types'

export function WorkDetail({work,onClose}:{work:Work|null;onClose:()=>void}){
 const {toggleLike,requireAuth}=useApp(),nav=useNavigate(),[msg,ctx]=message.useMessage()
 useEffect(()=>{if(!work)return;const close=(event:KeyboardEvent)=>{if(event.key==='Escape')onClose()};document.body.style.overflow='hidden';window.addEventListener('keydown',close);return()=>{document.body.style.overflow='';window.removeEventListener('keydown',close)}},[work,onClose])
 if(!work)return null
 const go=(reference=false)=>{if(!requireAuth())return;onClose();nav(`/create?prompt=${encodeURIComponent(work.prompt)}${reference?'&reference=1':''}`)}
 const copy=async()=>{await navigator.clipboard?.writeText(work.prompt);msg.success('提示词已复制')}
 const media=work.image_url||work.video_url||'',author=work.user?.name||'用户'
 return <div className="work-detail-overlay" role="dialog" aria-modal="true" aria-label="作品详情">{ctx}
  <div className="work-detail-preview"><button className="work-detail-close" aria-label="关闭" onClick={onClose}><CloseOutlined/></button>{work.type==='video'?<video src={media} controls poster={work.image_url}/>:<img src={media} alt={work.title}/>}</div>
  <aside className="work-detail-panel">
   <header className="work-detail-author"><span>{author.slice(0,1)}</span><div><b>{author}</b><small>{new Date(work.created_at).toLocaleString('zh-CN')}</small></div><button aria-label="喜欢" onClick={()=>void toggleLike(work.id)}>{work.is_liked?<HeartFilled className="text-[#ef7893]"/>:<HeartOutlined/>} {work.likes_count}</button></header>
   <h1>{work.title}</h1>
   <div className="work-detail-tags"><span>#{work.category}</span><span>#{work.type==='video'?'视频':'灵感'}</span><span>#{work.ratio}</span></div>
   <section className="work-detail-prompt"><div><span>{work.type==='video'?'视频提示词':'图片提示词'}</span><button aria-label="复制提示词" title="复制提示词" onClick={copy}><CopyOutlined/></button></div><p>{work.prompt}</p></section>
   <section className="work-detail-params"><span>生成参数 · {work.type==='video'?'视频':'图片'}</span><div><b>{work.ratio}</b><i>·</i><b>{work.type==='video'?'720P':'2K'}</b><i>·</i><b>{work.type==='video'?'Seedance':'GPT Image2'}</b></div></section>
   <div className="work-detail-meta"><span>{work.views_count} 浏览</span><span>{work.collects_count} 收藏</span></div>
   <footer className="work-detail-actions"><button onClick={()=>void toggleLike(work.id)}>{work.is_liked?'已喜欢':'喜欢'}</button><button onClick={()=>go(false)}>做同款</button><button onClick={()=>go(true)}>用作参考图</button></footer>
  </aside>
 </div>
}
