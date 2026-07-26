import { CloseOutlined, KeyOutlined, ReloadOutlined } from '@ant-design/icons'
import { Button, Form, Input, Modal, Spin, message } from 'antd'
import { useEffect, useState } from 'react'
import type { CreditPackage } from '../types'
import { apiClient } from '../utils/api'
import { useApp } from '../store/AppStore'

type Mode = 'packages' | 'redeem'
export function CreditModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { balance, refreshProfile, requireAuth } = useApp()
  const [mode, setMode] = useState<Mode>('packages')
  const [packages, setPackages] = useState<CreditPackage[]>([])
  const [loading, setLoading] = useState(false)
  const [redeeming, setRedeeming] = useState(false)
  const [form] = Form.useForm<{ code: string }>()
  useEffect(() => {
    if (!open) return
    setMode('packages'); setLoading(true)
    apiClient.creditPackages().then(setPackages).catch(error => message.error(error instanceof Error ? error.message : '加载套餐失败')).finally(() => setLoading(false))
  }, [open])
  const redeem = async ({ code }: { code: string }) => {
	if (!requireAuth()) return
    setRedeeming(true)
    try { const result = await apiClient.user.redeemCode(code.trim()); await refreshProfile(); form.resetFields(); message.success(`兑换成功，已到账 ${result.amount} 灵感值`); onClose() } catch (error) { message.error(error instanceof Error ? error.message : '兑换失败') } finally { setRedeeming(false) }
  }
  return <Modal open={open} onCancel={onClose} footer={null} closable={false} width={mode === 'packages' ? 1010 : 520} className="credit-modal" destroyOnHidden>
    <button className="credit-close" onClick={onClose} aria-label="关闭"><CloseOutlined /></button>
    {mode === 'packages' ? <section className="credit-packages"><h2>获取灵感，开启 AI 创作</h2><p>选择适合你的套餐，或直接 <button onClick={() => setMode('redeem')}>密钥兑换</button></p>{loading ? <div className="credit-loading"><Spin /></div> : <div className="credit-package-grid">{packages.map(item => <article className={`credit-package ${item.is_hot ? 'featured' : ''}`} key={item.id}>{item.is_hot && <span className="credit-badge">性价比最高</span>}<h3>{item.name}</h3><strong><small>¥</small>{item.price.toFixed(2)}</strong><span className="credit-points"><ReloadOutlined /> {item.points} 灵感</span><p>{item.note || '用于 AI 图片和视频生成'}</p><Button type={item.is_hot ? 'primary' : 'default'} block disabled={!item.purchase_url} onClick={() => window.open(item.purchase_url, '_blank', 'noopener,noreferrer')}>{item.purchase_url ? '前往购买' : '暂未配置链接'}</Button></article>)}</div>}<div className="credit-footer">已有密钥？ <button onClick={() => setMode('redeem')}><KeyOutlined /> 立即兑换</button></div></section> : <section className="credit-redeem"><div className="credit-mark"><ReloadOutlined /></div><h2>兑换灵感</h2><p>输入密钥兑换灵感，灵感可用于生成 AI 图像</p><div className="credit-balance">当前余额 <b>{balance}</b><ReloadOutlined /></div><Form form={form} layout="vertical" onFinish={redeem}><Form.Item name="code" label="卡密" rules={[{ required: true, message: '请输入兑换码' }]}><Input placeholder="例如：AGI-XXXX-XXXX-XXXXXXXX" autoCapitalize="characters" /></Form.Item><Form.Item><Button type="primary" htmlType="submit" block loading={redeeming}>兑换</Button></Form.Item></Form><div className="credit-divider">还没有密钥？</div><Button className="credit-back" onClick={() => setMode('packages')} icon={<ReloadOutlined />}>查看套餐与价格</Button></section>}
  </Modal>
}
