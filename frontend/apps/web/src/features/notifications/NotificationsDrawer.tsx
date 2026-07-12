import { Drawer, List, Tag } from 'antd'

const notices=[
 {title:'欢迎来到潮汐 AI',content:'新用户体验灵感值已到账，开始你的第一次创作。',time:'刚刚',type:'系统'},
 {title:'Flux Vision 2 已上线',content:'中文文字、人物细节与复杂构图能力全面提升。',time:'2 小时前',type:'产品'},
 {title:'本周创作挑战',content:'以「城市之外」为主题发布作品，赢取灵感值奖励。',time:'昨天',type:'活动'},
]
export function NotificationsDrawer({open,onClose}:{open:boolean;onClose:()=>void}){
 return <Drawer title="通知中心" width={420} open={open} onClose={onClose} extra={<span className="text-xs text-[#7f899b]">全部已读</span>}><List dataSource={notices} renderItem={item=><List.Item className="!items-start"><List.Item.Meta avatar={<span className="notice-dot"/>} title={<div className="flex justify-between gap-4"><b>{item.title}</b><small className="font-normal text-[#717b8d]">{item.time}</small></div>} description={<><p className="mb-2 mt-1 text-[#9ba4b5]">{item.content}</p><Tag bordered={false}>{item.type}</Tag></>}/></List.Item>}/></Drawer>
}
