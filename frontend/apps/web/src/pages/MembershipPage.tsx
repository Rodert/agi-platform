import { Empty } from 'antd'
import { useApp } from '../store/AppStore'

export function MembershipPage() {
  const { balance } = useApp()
  return <main className="page"><h1 className="page-title">我的账户</h1><p className="page-subtitle">查看灵感值余额与账户记录</p><section className="surface mt-6 p-6"><span className="text-[#8f98aa]">当前灵感值</span><div className="mt-2 text-4xl font-bold text-[#54d6cf]">{balance}</div></section><section className="surface mt-5 p-6"><h2 className="mt-0 text-lg">账户记录</h2><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="账户流水接口尚未开放"/></section></main>
}
