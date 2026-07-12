import { Dropdown, type MenuProps } from 'antd'
import type { CreationMode } from './ModeSwitcher'

const labels={image:'🎨 图片生成',video:'🎬 视频生成'} as const
const descriptions={image:'智能美学提升',video:'首尾帧精准控制'} as const

export function CreationModeDropdown({mode,onChange}:{mode:CreationMode;onChange?:(mode:CreationMode)=>void}){
 const items:MenuProps['items']=(Object.keys(labels) as CreationMode[]).map(key=>({key,label:<span className="creation-mode-option"><b>{labels[key].slice(0,2)}</b><span>{labels[key].slice(3)}<small>{descriptions[key]}</small></span></span>}))
 return <Dropdown menu={{items,selectedKeys:[mode],onClick:({key})=>onChange?.(key as CreationMode)}} trigger={['click']} placement="bottomLeft" overlayClassName="creation-mode-menu"><button className="generator-type">{labels[mode]}⌄</button></Dropdown>
}
