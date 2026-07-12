export type CreationMode='image'|'video'
const modes=[
 {key:'image',label:'图片生成',description:'智能美学提升',icon:'🎨'},
 {key:'video',label:'视频生成',description:'首尾帧精准控制',icon:'🎬'},
] as const

export function ModeSwitcher({mode,onChange,onTools}:{mode:CreationMode;onChange:(mode:CreationMode)=>void;onTools:()=>void}){return <div className="creation-shortcuts">{modes.map(item=><button key={item.key} className={mode===item.key?'active':''} onClick={()=>onChange(item.key)}><i>{item.icon}</i><span><b>{item.label}</b><small>{item.description}</small></span></button>)}<button onClick={onTools}><i>🧰</i><span><b>工具箱</b><small>实用创作助手</small></span></button></div>}
