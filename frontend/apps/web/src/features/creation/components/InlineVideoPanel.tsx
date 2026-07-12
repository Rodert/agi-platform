import { ArrowUpOutlined } from '@ant-design/icons'
import { Button, Input, Select, message } from 'antd'
import { useEffect, useState } from 'react'
import { useApp } from '../../../store/AppStore'
import { CreationModeDropdown } from './CreationModeDropdown'
import type { CreationMode } from './ModeSwitcher'
import { ModelParams, modelParamCost, modelParamDefaults } from './ModelParams'

export function InlineVideoPanel({onModeChange}:{onModeChange:(mode:CreationMode)=>void}){
 const [prompt,setPrompt]=useState(''),[model,setModel]=useState(''),[modelParams,setModelParams]=useState<Record<string,unknown>>({}),[msg,ctx]=message.useMessage(),{createTask,requireAuth,models}=useApp()
 const videoModels=models.filter(item=>item.type==='video'),current=videoModels.find(item=>item.name===model)||videoModels[0]
 useEffect(()=>{setModelParams(modelParamDefaults(current))},[current])
 const submit=async()=>{if(!requireAuth())return;if(!prompt.trim())return msg.warning('请输入视频画面与运动描述');if(!current)return msg.error('暂无可用视频模型');if(await createTask({prompt:prompt.trim(),type:'video',modelName:current.name,params:modelParams}))msg.success('视频任务已提交');else msg.error('任务提交失败')}
 return <>{ctx}<section className="surface generator-panel overflow-hidden"><div className="flex min-h-[108px] gap-3 p-4"><Input.TextArea autoSize={{minRows:3,maxRows:6}} variant="borderless" value={prompt} onChange={e=>setPrompt(e.target.value)} placeholder="输入文字，描述你想创作的画面内容、运动方式等。"/></div><div className="generator-toolbar"><CreationModeDropdown mode="video" onChange={onModeChange}/><Select variant="borderless" value={current?.name} placeholder="暂无可用模型" onChange={setModel} options={videoModels.map(item=>({label:item.display_name,value:item.name}))}/><ModelParams model={current} values={modelParams} onChange={setModelParams} compact/><span className="ml-auto">◆ {current?modelParamCost(current,modelParams):'-'}</span><Button type="primary" shape="circle" icon={<ArrowUpOutlined/>} disabled={!current} onClick={()=>void submit()}/></div></section></>
}
