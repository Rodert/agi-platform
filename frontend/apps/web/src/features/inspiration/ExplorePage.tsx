import { SearchOutlined, UploadOutlined } from '@ant-design/icons'
import { Button, Input, Segmented } from 'antd'
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { GeneratePanel, InlineVideoPanel, ModeSwitcher, type CreationMode } from '../creation'
import { WorkCard } from './components/WorkCard'
import { WorkDetail } from './components/WorkDetail'
import { categories } from '../../mock/data'
import { useApp } from '../../store/AppStore'
import type { Work } from '../../types'
import { PublishWorkModal } from '../works'

export function ExplorePage(){
 const {works,tasks,requireAuth,loadTasks}=useApp(),nav=useNavigate(),[searchParams]=useSearchParams()
 const promptFromWork=searchParams.get('prompt')||'', requestedMode=searchParams.get('mode')
 const [mode,setMode]=useState<CreationMode>(requestedMode==='video'?'video':'image'),[type,setType]=useState('全部'),[cat,setCat]=useState('全部'),[query,setQuery]=useState(''),[searchTerm,setSearchTerm]=useState(''),[searchFocused,setSearchFocused]=useState(false),[active,setActive]=useState<Work|null>(null),[publishOpen,setPublishOpen]=useState(false)
 const [columnCount,setColumnCount]=useState(()=>window.innerWidth<=800?2:5)
 const refill=promptFromWork?{prompt:promptFromWork,key:`${requestedMode}:${promptFromWork}`} : undefined
 useEffect(()=>{if(promptFromWork){setMode(requestedMode==='video'?'video':'image');document.querySelector('.inspiration-generator-zone')?.scrollIntoView({behavior:'smooth',block:'start'})}},[promptFromWork,requestedMode])
 useEffect(()=>{const update=()=>setColumnCount(window.innerWidth<=800?2:5);window.addEventListener('resize',update);return()=>window.removeEventListener('resize',update)},[])
 const list=useMemo(()=>works.filter(w=>(type==='全部'||(type==='图片'?w.type==='image':w.type==='video'))&&(cat==='全部'||w.category===cat)&&(!searchTerm||`${w.title}${w.prompt}`.toLowerCase().includes(searchTerm.toLowerCase()))),[works,type,cat,searchTerm])
 const columns=useMemo(()=>Array.from({length:columnCount},(_,index)=>list.filter((_,workIndex)=>workIndex%columnCount===index)),[columnCount,list])
 const search=()=>setSearchTerm(query.trim())
 return <main className="page inspiration-page">
  <div className="inspiration-generator-zone">
   <div className="theme-muted mb-2 flex items-center gap-2 text-xs"><span className="theme-accent">◉ 新品上线</span><span>GPT Image 2 已开放体验，中文与细节表现全面提升</span></div>
   {mode==='image'?<GeneratePanel refill={refill} onModeChange={setMode}/>:<InlineVideoPanel refill={refill} onModeChange={setMode}/>}<ModeSwitcher mode={mode} onChange={setMode} onTools={()=>nav('/tools')}/>
  </div>
  <section className="inspiration-showcase">
   <div className="discover-bar"><Segmented className="work-type-segmented" options={['全部','图片','视频']} value={type} onChange={v=>setType(String(v))}/><Input prefix={<SearchOutlined/>} suffix={(searchFocused||query)&&<Button className="search-submit" type="primary" size="small" icon={<SearchOutlined/>} onMouseDown={e=>e.preventDefault()} onClick={search}>搜索</Button>} placeholder="搜索标题或提示词" allowClear className="work-search-input !w-80 max-[800px]:!w-full" value={query} onFocus={()=>setSearchFocused(true)} onBlur={()=>setSearchFocused(false)} onPressEnter={search} onChange={e=>{setQuery(e.target.value);if(!e.target.value)setSearchTerm('')}}/><Button icon={<UploadOutlined/>} onClick={()=>{if(requireAuth()){void loadTasks();setPublishOpen(true)}}}>发布作品</Button></div>
   <div className="mb-3 flex gap-2 overflow-auto pb-1">{categories.map(x=><button key={x} onClick={()=>setCat(x)} className={`category-filter whitespace-nowrap rounded-full px-3 py-1 text-xs ${cat===x?'is-active':''}`}>{x}</button>)}</div>
   {list.length?<div className="masonry">{columns.map((column,index)=><div className="masonry-column" key={index}>{column.map(w=><WorkCard key={w.id} work={w} onOpen={setActive}/>)}</div>)}</div>:<div className="theme-muted surface py-20 text-center">没有找到相关作品</div>}
  </section>
  <WorkDetail work={active} onClose={()=>setActive(null)}/>
  <PublishWorkModal open={publishOpen} tasks={tasks} onClose={()=>setPublishOpen(false)} onPublished={loadTasks}/>
 </main>
}
