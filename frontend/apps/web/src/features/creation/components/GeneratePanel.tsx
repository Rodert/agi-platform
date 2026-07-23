import { ArrowUpOutlined, HighlightOutlined } from '@ant-design/icons'
import { Button, Input, Select, Tooltip, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useApp } from '../../../store/AppStore'
import type { CreationMode } from './ModeSwitcher'
import { CreationModeDropdown } from './CreationModeDropdown'
import { ModelParams, modelParamCost, modelParamDefaults } from './ModelParams'
import { ReferenceImagePicker } from './ReferenceImagePicker'
import { apiClient } from '../../../utils/api'

const modelLogos:Record<string,{src?:string;mark?:string;className:string}>={
 'GPT Image2':{src:'/common-logo/chatgpt-logo.png',className:'gpt'},
 'Seedream':{src:'/common-logo/jm-logo.png',className:'seedream'},
 'Nanobanana':{mark:'🍌',className:'banana'},
 'GPT Image2 备用':{src:'/common-logo/chatgpt-logo.png',className:'gpt'},
}
const modelLabel=(label:string)=>{const logo=modelLogos[label];return <span className="model-option-label">{logo&&<i className={`model-logo ${logo.className}`}>{logo.src?<img src={logo.src} alt=""/>:logo.mark}</i>}<span>{label}</span></span>}

export function GeneratePanel({onModeChange,refill}:{onModeChange?:(mode:CreationMode)=>void;refill?:{prompt:string;key:number}}={}){
 const [queryParams]=useSearchParams(),navigate=useNavigate(),[prompt,setPrompt]=useState(''),[references,setReferences]=useState<string[]>([]),[model,setModel]=useState(''),[modelParams,setModelParams]=useState<Record<string,unknown>>({}),[optimizing,setOptimizing]=useState(false),[msg,ctx]=message.useMessage(),{createTask,requireAuth,models}=useApp()
 const imageModels=models.filter(item=>item.type==='image')
 useEffect(()=>{const p=queryParams.get('prompt');if(p)setPrompt(p)},[queryParams])
 useEffect(()=>{if(refill){setPrompt(refill.prompt);requestAnimationFrame(()=>document.querySelector<HTMLTextAreaElement>('.generation-composer textarea')?.focus())}},[refill])
 useEffect(()=>{if(!model&&imageModels[0])setModel(imageModels[0].name)},[imageModels,model])
 const current=imageModels.find(x=>x.name===model)
 useEffect(()=>{setModelParams(modelParamDefaults(current))},[current])
 const optimize=async()=>{if(!requireAuth())return;if(!prompt.trim())return msg.warning('先输入提示词');if(!current)return msg.warning('请先选择图片模型');setOptimizing(true);try{const result=await apiClient.generation.optimizePrompt({prompt:prompt.trim(),target_type:'image',target_model_name:current.name,params:modelParams});setPrompt(result.prompt);msg.success(result.credit_cost>0?`提示词已优化，消耗 ${result.credit_cost} 灵感值`:'提示词已优化')}catch(error){msg.error(error instanceof Error?error.message:'提示词优化失败')}finally{setOptimizing(false)}}
 const submit=async()=>{if(!requireAuth())return;if(!prompt.trim())return msg.warning('先描述你想生成的画面');if(!current)return msg.error('暂无可用图片模型');const ok=await createTask({prompt:prompt.trim(),type:'image',modelName:current.name,params:modelParams,referenceImage:references[0]});if(!ok)return msg.error('任务提交失败');setPrompt('');setReferences([]);navigate('/create')}
 return <>{ctx}<section className="surface generator-panel overflow-hidden"><div className="flex min-h-[108px] gap-3 p-4"><ReferenceImagePicker value={references} onChange={setReferences}/><Input.TextArea autoSize={{minRows:3,maxRows:6}} variant="borderless" value={prompt} onChange={e=>setPrompt(e.target.value)} placeholder="描述你想生成的画面，AI 将为你生成精美图片..." className="!text-base"/><Tooltip title="优化提示词"><Button aria-label="优化提示词" type="text" shape="circle" icon={<HighlightOutlined/>} onClick={()=>void optimize()} loading={optimizing} disabled={optimizing}/></Tooltip></div><div className="generator-toolbar"><CreationModeDropdown mode="image" onChange={onModeChange}/><Select aria-label="图片模型" popupMatchSelectWidth={220} variant="borderless" value={model||undefined} placeholder="暂无可用模型" onChange={setModel} options={imageModels.map(x=>({label:modelLabel(x.display_name),value:x.name}))} className="image-model-select"/><ModelParams model={current} values={modelParams} onChange={setModelParams} compact/><span className="ml-auto text-[#8d8da2]">◆ <b className="text-[#b3b3c2]">{current?modelParamCost(current,modelParams):'-'}</b></span><Button type="primary" shape="circle" icon={<ArrowUpOutlined/>} onClick={()=>void submit()} disabled={!current}/></div></section></>
}
