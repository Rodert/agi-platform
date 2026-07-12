import { LockOutlined, MailOutlined } from '@ant-design/icons'
import { Button, Checkbox, Form, Input, Modal, Segmented, Tabs, message } from 'antd'
import { useEffect, useState } from 'react'

interface LoginDialogProps { open:boolean; onClose:()=>void; onLogin:(email:string,password:string)=>Promise<boolean> }
type AuthMode='login'|'register'
type AuthMethod='code'|'password'

export function LoginDialog({open,onClose,onLogin}:LoginDialogProps){
 const [mode,setMode]=useState<AuthMode>('login'),[method,setMethod]=useState<AuthMethod>('code'),[countdown,setCountdown]=useState(0)
 const [form]=Form.useForm();const [msg,ctx]=message.useMessage()
 useEffect(()=>{if(!countdown)return;const timer=window.setInterval(()=>setCountdown(v=>Math.max(0,v-1)),1000);return()=>window.clearInterval(timer)},[countdown])
 useEffect(()=>{if(open){setMode('login');setMethod('code');setCountdown(0);form.resetFields()}},[open,form])
 const changeMode=(key:string)=>{setMode(key as AuthMode);setMethod('code');setCountdown(0);form.resetFields()}
 const sendCode=async()=>{try{await form.validateFields(['email']);setCountdown(60);msg.success('验证码已发送，请查收邮箱')}catch{/* Form displays the email error. */}}
 const [submitting,setSubmitting]=useState(false)
 const submit=async(values:{email:string;password?:string})=>{
  if(mode!=='login'||method!=='password'){msg.info('当前环境请使用邮箱密码登录');return}
  setSubmitting(true)
  const success=await onLogin(values.email,values.password||'')
  if(!success)msg.error('邮箱或密码错误')
  setSubmitting(false)
 }
 const action=mode==='login'?'登录':'注册'
 const isRegister=mode==='register'
 return <Modal className="auth-modal" width={440} title={null} open={open} footer={null} onCancel={onClose} destroyOnHidden>{ctx}
  <div className="auth-heading"><img src="/logo/website-logo.png" alt=""/><div><h2>潮汐 AI</h2><p>{mode==='login'?'欢迎回来，继续你的创作':'创建账户，开始你的创作'}</p></div></div>
  <Tabs activeKey={mode} onChange={changeMode} centered items={[{key:'login',label:'登录'},{key:'register',label:'注册'}]}/>
  {!isRegister&&<Segmented block value={method} onChange={value=>{setMethod(value as AuthMethod);form.resetFields(['code','password','confirmPassword'])}} options={[{label:'邮箱验证码',value:'code'},{label:'邮箱密码',value:'password'}]}/>} 
  <Form form={form} className="auth-form" layout="vertical" onFinish={submit} requiredMark={false}>
   <Form.Item label="邮箱" name="email" rules={[{required:true,message:'请输入邮箱地址'},{type:'email',message:'请输入有效的邮箱地址'}]}><Input size="large" prefix={<MailOutlined/>} autoComplete="email" placeholder="name@example.com"/></Form.Item>
   {(isRegister||method==='code')&&<Form.Item label="验证码" name="code" rules={[{required:true,message:'请输入验证码'},{len:6,message:'请输入 6 位验证码'}]}><Input size="large" inputMode="numeric" maxLength={6} placeholder="请输入 6 位验证码" suffix={<Button type="link" size="small" disabled={countdown>0} onClick={sendCode}>{countdown?`${countdown}s 后重发`:'获取验证码'}</Button>}/></Form.Item>}
   {(isRegister||method==='password')&&<>
    <Form.Item label="密码" name="password" rules={[{required:true,message:'请输入密码'},{min:8,message:'密码至少 8 位'}]}><Input.Password size="large" prefix={<LockOutlined/>} autoComplete={mode==='login'?'current-password':'new-password'} placeholder="至少 8 位密码"/></Form.Item>
    {isRegister&&<Form.Item label="确认密码" name="confirmPassword" dependencies={['password']} rules={[{required:true,message:'请再次输入密码'},({getFieldValue})=>({validator(_,value){return !value||getFieldValue('password')===value?Promise.resolve():Promise.reject(new Error('两次输入的密码不一致'))}})]}><Input.Password size="large" prefix={<LockOutlined/>} autoComplete="new-password" placeholder="再次输入密码"/></Form.Item>}
   </>}
   <Form.Item className="auth-consent" name="consent" valuePropName="checked" rules={[{validator:(_,checked)=>checked?Promise.resolve():Promise.reject(new Error('请先阅读并同意用户协议和隐私政策'))}]}>
    <Checkbox>我已阅读并同意 <a href="/user-agreement" target="_blank" rel="noreferrer" onClick={event=>event.stopPropagation()}>《用户协议》</a>和<a href="/privacy-policy" target="_blank" rel="noreferrer" onClick={event=>event.stopPropagation()}>《隐私政策》</a></Checkbox>
   </Form.Item>
   <Button className="auth-submit" block size="large" type="primary" htmlType="submit" loading={submitting}>{action}</Button>
  </Form>
 </Modal>
}
