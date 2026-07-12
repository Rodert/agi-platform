import { CheckOutlined, CustomerServiceOutlined, TeamOutlined } from '@ant-design/icons'
import { Button, Modal, Segmented, message } from 'antd'
import { useState } from 'react'

export function CommunityDialog({open,onClose}:{open:boolean;onClose:()=>void}){
 const [type,setType]=useState('微信社群'),[msg,ctx]=message.useMessage()
 return <>{ctx}<Modal title="加入潮汐创作者社群" open={open} footer={null} onCancel={onClose}><Segmented block options={['微信社群','QQ 社群','创作者社区']} value={type} onChange={v=>setType(String(v))}/><div className="community-panel"><div className="community-mark"><TeamOutlined/></div><h3>{type}</h3><p>{type==='创作者社区'?'交流提示词、工作流与模型体验，参与每周创作挑战。':'获取模型资讯、创作技巧、活动奖励和问题答疑。'}</p><div className="community-benefits"><span><CheckOutlined/> 新模型内测</span><span><CheckOutlined/> 每周创作挑战</span><span><CheckOutlined/> 专属问题答疑</span></div><Button type="primary" icon={<CustomerServiceOutlined/>} onClick={()=>msg.success(`${type}加入信息已复制`)}>获取加入方式</Button></div></Modal></>
}
