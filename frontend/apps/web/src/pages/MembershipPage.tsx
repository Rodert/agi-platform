import { Button, Empty, Form, Input, Pagination, Table, Tag, message } from 'antd'
import { useEffect, useState } from 'react'
import type { Ledger } from '../types'
import { apiClient } from '../utils/api'
import { useApp } from '../store/AppStore'

export function MembershipPage() {
  const { balance, user, refreshProfile } = useApp()
  const [items, setItems] = useState<Ledger[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [redeeming, setRedeeming] = useState(false)
  const [form] = Form.useForm<{ code: string }>()

  const redeemCode = async ({ code }: { code: string }) => {
    setRedeeming(true)
    try {
      const result = await apiClient.user.redeemCode(code.trim())
      message.success(`兑换成功，已到账 ${result.amount} 灵感值`)
      form.resetFields()
      await refreshProfile()
      setPage(1)
    } catch (error) {
      message.error(error instanceof Error ? error.message : '兑换失败')
    } finally { setRedeeming(false) }
  }

  useEffect(() => {
    if (!user) { setItems([]); setTotal(0); return }
    let active = true
    setLoading(true)
    apiClient.user.getCreditLedgers({ page, page_size: 20 })
      .then(result => { if (active) { setItems(result.list); setTotal(result.total) } })
      .catch(() => { if (active) { setItems([]); setTotal(0) } })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [page, user])

  const columns = [
    { title: '说明', dataIndex: 'title', key: 'title' },
    { title: '变动', key: 'amount', width: 100, render: (_: unknown, row: Ledger) => <span className={row.type === 'income' ? 'theme-accent' : 'text-[#ff7a7a]'}>{row.type === 'income' ? '+' : '-'}{row.amount}</span> },
    { title: '余额', dataIndex: 'balance_after', key: 'balance_after', width: 90 },
    { title: '类型', key: 'type', width: 90, render: (_: unknown, row: Ledger) => <Tag color={row.type === 'income' ? 'cyan' : 'red'}>{row.type === 'income' ? '收入' : '支出'}</Tag> },
    { title: '时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  ]

  return <main className="page"><h1 className="page-title">我的账户</h1><p className="page-subtitle">查看灵感值余额与账户记录</p><section className="surface mt-6 p-6"><span className="theme-muted">当前灵感值</span><div className="theme-accent mt-2 text-4xl font-bold">{balance}</div></section><section className="surface mt-5 p-6"><h2 className="mt-0 text-lg">兑换灵感值</h2><Form form={form} layout="inline" onFinish={redeemCode}><Form.Item name="code" rules={[{ required: true, message: '请输入兑换码' }]}><Input className="min-w-64" placeholder="输入兑换码" autoCapitalize="characters" /></Form.Item><Form.Item><Button type="primary" htmlType="submit" loading={redeeming} disabled={!user}>立即兑换</Button></Form.Item></Form>{!user && <p className="theme-muted mb-0 text-sm">登录后可兑换灵感值。</p>}</section><section className="surface mt-5 p-6"><h2 className="mt-0 text-lg">账户记录</h2>{user ? <><Table className="mt-4" columns={columns} dataSource={items} rowKey="id" loading={loading} pagination={false} locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无账户流水" /> }} scroll={{ x: 680 }} /><div className="mt-4 flex justify-end"><Pagination current={page} pageSize={20} total={total} showSizeChanger={false} onChange={setPage} /></div></> : <Empty className="mt-6" image={Empty.PRESENTED_IMAGE_SIMPLE} description="登录后查看账户流水" />}</section></main>
}
