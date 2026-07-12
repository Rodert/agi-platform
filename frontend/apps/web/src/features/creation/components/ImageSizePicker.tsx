import { Popover } from 'antd'
import { useState } from 'react'

const ratios=['auto','3:2','1:1','2:3','5:4','4:5','16:9','9:16','21:9','3:4','4:3']
const ratioSize:Record<string,[number,number]>={auto:[16,16],'3:2':[18,12],'1:1':[16,16],'2:3':[12,18],'5:4':[18,14],'4:5':[14,18],'16:9':[20,11],'9:16':[11,20],'21:9':[22,9],'3:4':[13,18],'4:3':[18,13]}

export function ImageSizePicker({ratio,resolution,onChange}:{ratio:string;resolution:string;onChange:(ratio:string,resolution:string)=>void}){
 const [open,setOpen]=useState(false)
 const content=<div className="size-popover"><label>选择比例</label><div className="ratio-grid">{ratios.map(item=>{const [w,h]=ratioSize[item];return <button key={item} className={ratio===item?'active':''} onClick={()=>onChange(item,resolution)}><i style={{width:w,height:h}}/><span>{item}</span></button>})}</div><label>选择分辨率</label><div className="resolution-tabs">{['1K','2K','4K'].map(item=><button key={item} className={resolution===item?'active':''} onClick={()=>{onChange(ratio,item);setOpen(false)}}><b>{item}</b><small>4 ◆</small></button>)}</div></div>
 return <Popover open={open} onOpenChange={setOpen} trigger="click" placement="bottomLeft" arrow={false} content={content} overlayClassName="size-popover-wrap"><button className={`parameter-trigger ${open?'active':''}`}>▢ {ratio} <span>{resolution}</span>⌄</button></Popover>
}
