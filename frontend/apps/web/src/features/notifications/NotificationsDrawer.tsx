import { Drawer, Empty, List, Spin, Tag } from 'antd'
import { useEffect, useState } from 'react'
import type { Announcement } from '../../types'
import { apiClient } from '../../utils/api'

export function NotificationsDrawer({open,onClose}:{open:boolean;onClose:()=>void}){
 const [items,setItems]=useState<Announcement[]>([]),[loading,setLoading]=useState(false)
 useEffect(()=>{if(!open)return;setLoading(true);apiClient.notifications.list({page:1,page_size:50}).then(result=>setItems(result.list)).catch(error=>console.error('加载通知失败:',error)).finally(()=>setLoading(false))},[open])
 return <Drawer title="通知中心" width={420} open={open} onClose={onClose}>{loading?<div className="grid min-h-40 place-items-center"><Spin/></div>:items.length?<List dataSource={items} renderItem={item=><List.Item className="!items-start"><List.Item.Meta avatar={<span className="notice-dot"/>} title={<div className="flex justify-between gap-4"><b>{item.title}</b><small className="font-normal text-[#717b8d]">{new Date(item.published_at||item.created_at).toLocaleString('zh-CN')}</small></div>} description={<><p className="mb-2 mt-1 whitespace-pre-wrap text-[#9ba4b5]">{item.content}</p><Tag bordered={false}>{item.category==='activity'?'活动':item.category==='product'?'产品':'系统'}</Tag></>}/></List.Item>}/>:<Empty description="暂无通知" image={Empty.PRESENTED_IMAGE_SIMPLE}/>}</Drawer>
}
