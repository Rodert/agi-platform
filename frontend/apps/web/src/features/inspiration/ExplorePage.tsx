import { SearchOutlined, UploadOutlined } from '@ant-design/icons'
import { Button, Input, Segmented } from 'antd'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { GeneratePanel, InlineVideoPanel, ModeSwitcher, type CreationMode } from '../creation'
import { WorkCard } from './components/WorkCard'
import { WorkDetail } from './components/WorkDetail'
import { categories } from '../../mock/data'
import { useApp } from '../../store/AppStore'
import type { Work } from '../../types'

export function ExplorePage(){
 const {works,requireAuth}=useApp(),nav=useNavigate()
 const [mode,setMode]=useState<CreationMode>('image'),[type,setType]=useState('全部'),[cat,setCat]=useState('全部'),[query,setQuery]=useState(''),[searchTerm,setSearchTerm]=useState(''),[searchFocused,setSearchFocused]=useState(false),[active,setActive]=useState<Work|null>(null)
 const list=useMemo(()=>works.filter(w=>(type==='全部'||(type==='图片'?w.type==='image':w.type==='video'))&&(cat==='全部'||w.category===cat)&&(!searchTerm||`${w.title}${w.prompt}`.toLowerCase().includes(searchTerm.toLowerCase()))),[works,type,cat,searchTerm])
 const search=()=>setSearchTerm(query.trim())
 return <main className="page inspiration-page">
  <div className="inspiration-generator-zone">
   <div className="mb-2 flex items-center gap-2 text-xs text-[#8f98aa]"><span className="text-[#54d6cf]">◉ 新品上线</span><span>GPT Image 2 已开放体验，中文与细节表现全面提升</span></div>
   {mode==='image'?<GeneratePanel onModeChange={setMode}/>:<InlineVideoPanel onModeChange={setMode}/>}<ModeSwitcher mode={mode} onChange={setMode} onTools={()=>nav('/tools')}/>
  </div>
  <section className="inspiration-showcase">
   <div className="discover-bar"><Segmented className="work-type-segmented" options={['全部','图片','视频']} value={type} onChange={v=>setType(String(v))}/><Input prefix={<SearchOutlined/>} suffix={(searchFocused||query)&&<Button className="search-submit" type="primary" size="small" icon={<SearchOutlined/>} onMouseDown={e=>e.preventDefault()} onClick={search}>搜索</Button>} placeholder="搜索标题或提示词" allowClear className="work-search-input !w-80 max-[800px]:!w-full" value={query} onFocus={()=>setSearchFocused(true)} onBlur={()=>setSearchFocused(false)} onPressEnter={search} onChange={e=>{setQuery(e.target.value);if(!e.target.value)setSearchTerm('')}}/><Button icon={<UploadOutlined/>} onClick={()=>requireAuth()}>发布作品</Button></div>
   <div className="mb-3 flex gap-2 overflow-auto pb-1">{categories.map(x=><button key={x} onClick={()=>setCat(x)} className={`whitespace-nowrap rounded-full border px-3 py-1 text-xs ${cat===x?'border-[#54d6cf] bg-[#173033] text-[#bff9f5]':'border-[#2a303e] bg-[#131720] text-[#929bad]'}`}>{x}</button>)}</div>
   {list.length?<div className="masonry">{list.map(w=><WorkCard key={w.id} work={w} onOpen={setActive}/>)}</div>:<div className="surface py-20 text-center text-[#8e97a9]">没有找到相关作品</div>}
  </section>
  <WorkDetail work={active} onClose={()=>setActive(null)}/>
 </main>
}
