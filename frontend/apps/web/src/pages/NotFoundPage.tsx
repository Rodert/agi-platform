import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'
export function NotFoundPage(){const nav=useNavigate();return <main className="page"><Result status="404" title="页面不存在" subTitle="这个创作入口可能已经移动" extra={<Button type="primary" onClick={()=>nav('/')}>返回灵感广场</Button>}/></main>}
