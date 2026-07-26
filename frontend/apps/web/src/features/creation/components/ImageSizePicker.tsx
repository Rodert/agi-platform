import { Popover } from 'antd'
import { useState } from 'react'

export const GPT_IMAGE_2_SIZE_MAP={
 '1K':{'1:1':'1024x1024','4:3':'1024x768','3:4':'768x1024','3:2':'1152x768','2:3':'768x1152','5:4':'960x768','4:5':'768x960','16:9':'1280x720','9:16':'720x1280','2:1':'1280x640','1:2':'640x1280','21:9':'1344x576','9:21':'576x1344','3:1':'1536x512','1:3':'512x1536'},
 '2K':{'1:1':'2048x2048','4:3':'2048x1536','3:4':'1536x2048','3:2':'2304x1536','2:3':'1536x2304','5:4':'2000x1600','4:5':'1600x2000','16:9':'2560x1440','9:16':'1440x2560','2:1':'2048x1024','1:2':'1024x2048','21:9':'2688x1152','9:21':'1152x2688','3:1':'3072x1024','1:3':'1024x3072'},
 '4K':{'1:1':'2880x2880','4:3':'3200x2400','3:4':'2400x3200','3:2':'3456x2304','2:3':'2304x3456','16:9':'3840x2160','9:16':'2160x3840'},
} as const

type Resolution=keyof typeof GPT_IMAGE_2_SIZE_MAP
type Ratio=string
const ratioSize:Record<string,[number,number]>={'1:1':[16,16],'4:3':[18,14],'3:4':[14,18],'3:2':[18,12],'2:3':[12,18],'5:4':[18,14],'4:5':[14,18],'16:9':[20,11],'9:16':[11,20],'2:1':[20,10],'1:2':[10,20],'21:9':[22,9],'9:21':[9,22],'3:1':[22,8],'1:3':[8,22]}
const resolutionCost:Record<Resolution,string>={'1K':'4','2K':'5','4K':'6'}

export function ImageSizePicker({ratio,resolution,onChange}:{ratio:string;resolution:string;onChange:(ratio:string,resolution:string)=>void}){
 const [open,setOpen]=useState(false)
 const currentResolution=(resolution in GPT_IMAGE_2_SIZE_MAP?resolution:'1K') as Resolution
 const ratios=Object.keys(GPT_IMAGE_2_SIZE_MAP[currentResolution])
 const selectResolution=(next:Resolution)=>{const nextRatio=GPT_IMAGE_2_SIZE_MAP[next][ratio as keyof typeof GPT_IMAGE_2_SIZE_MAP[typeof next]]?ratio:(Object.keys(GPT_IMAGE_2_SIZE_MAP[next])[0] as Ratio);onChange(nextRatio,next);setOpen(false)}
 const content=<div className="size-popover"><label>选择比例</label><div className="ratio-grid">{ratios.map(item=>{const [w,h]=ratioSize[item];return <button key={item} className={ratio===item?'active':''} onClick={()=>onChange(item,currentResolution)}><i style={{width:w,height:h}}/><span>{item}</span></button>})}</div><label>选择分辨率</label><div className="resolution-tabs">{(Object.keys(GPT_IMAGE_2_SIZE_MAP) as Resolution[]).map(item=><button key={item} className={currentResolution===item?'active':''} onClick={()=>selectResolution(item)}><b>{item}</b><small>◆ {resolutionCost[item]}</small></button>)}</div><small className="size-value">{GPT_IMAGE_2_SIZE_MAP[currentResolution][ratio as keyof typeof GPT_IMAGE_2_SIZE_MAP[typeof currentResolution]]}</small></div>
 return <Popover open={open} onOpenChange={setOpen} trigger="click" placement="bottomLeft" arrow={false} content={content} overlayClassName="size-popover-wrap"><button className={`parameter-trigger ${open?'active':''}`}>▢ {ratio} <span>{currentResolution}</span>⌄</button></Popover>
}
