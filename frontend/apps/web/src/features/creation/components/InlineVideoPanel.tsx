import { ArrowUpOutlined, HighlightOutlined } from '@ant-design/icons'
import { Button, Input, Select, Tooltip, message } from 'antd'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useApp } from '../../../store/AppStore'
import { CreationModeDropdown } from './CreationModeDropdown'
import type { CreationMode } from './ModeSwitcher'
import { ModelParams, modelParamCost, modelParamDefaults } from './ModelParams'
import { ReferenceImagePicker } from './ReferenceImagePicker'
import { apiClient } from '../../../utils/api'

export function InlineVideoPanel({onModeChange,refill}:{onModeChange:(mode:CreationMode)=>void;refill?:{prompt:string;key:string|number}}){
 const [prompt,setPrompt]=useState(''),navigate=useNavigate(),[references,setReferences]=useState<string[]>([]),[model,setModel]=useState(''),[modelParams,setModelParams]=useState<Record<string,unknown>>({}),[optimizing,setOptimizing]=useState(false),[msg,ctx]=message.useMessage(),{createTask,requireAuth,models}=useApp()
 const videoModels=models.filter(item=>item.type==='video'),current=videoModels.find(item=>item.name===model)||videoModels[0]
 useEffect(()=>{if(refill){setPrompt(refill.prompt);requestAnimationFrame(()=>document.querySelector<HTMLTextAreaElement>('.inspiration-generator-zone textarea, .generation-composer textarea')?.focus())}},[refill])
 useEffect(()=>{setModelParams(modelParamDefaults(current))},[current])
 const optimize=async()=>{if(!requireAuth())return;if(!prompt.trim())return msg.warning('先输入提示词');if(!current)return msg.warning('请先选择视频模型');setOptimizing(true);try{const result=await apiClient.generation.optimizePrompt({prompt:prompt.trim(),target_type:'video',target_model_name:current.name,params:modelParams});setPrompt(result.prompt);msg.success(result.credit_cost>0?`提示词已优化，消耗 ${result.credit_cost} 灵感值`:'提示词已优化')}catch(error){msg.error(error instanceof Error?error.message:'提示词优化失败')}finally{setOptimizing(false)}}
 const submit=async()=>{if(!requireAuth())return;if(!prompt.trim())return msg.warning('请输入视频画面与运动描述');if(!current)return msg.error('暂无可用视频模型');const result=await createTask({prompt:prompt.trim(),type:'video',modelName:current.name,params:modelParams,referenceImages:references});if(result==='success'){setPrompt('');setReferences([]);navigate('/create')}else if(result==='failed')msg.error('任务提交失败')}
 return <>{ctx}<section className="surface generator-panel overflow-hidden"><div className="flex min-h-[108px] gap-3 p-4"><ReferenceImagePicker value={references} onChange={setReferences} max={3}/><Input.TextArea autoSize={{minRows:3,maxRows:6}} variant="borderless" value={prompt} onChange={e=>setPrompt(e.target.value)} placeholder="输入文字，描述你想创作的画面内容、运动方式等。"/><Tooltip title="优化提示词"><Button aria-label="优化提示词" type="text" shape="circle" icon={<HighlightOutlined/>} onClick={()=>void optimize()} loading={optimizing} disabled={optimizing}/></Tooltip></div><div className="generator-toolbar"><CreationModeDropdown mode="video" onChange={onModeChange}/><Select aria-label="视频模型" popupMatchSelectWidth={240} popupClassName="video-model-dropdown" variant="borderless" value={current?.name} placeholder="暂无可用模型" onChange={setModel} options={videoModels.map(item=>({label:<span className="video-model-option" title={item.display_name}>{item.display_name}</span>,value:item.name}))} className="video-model-select"/><ModelParams model={current} values={modelParams} onChange={setModelParams} compact/><span className="ml-auto">◆ {current?modelParamCost(current,modelParams):'-'}</span><Button type="primary" shape="circle" icon={<ArrowUpOutlined/>} disabled={!current} onClick={()=>void submit()}/></div></section></>
}
