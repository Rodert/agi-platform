import { LockOutlined, MailOutlined } from '@ant-design/icons'
import { Button, Checkbox, Form, Input, Modal, Segmented, message } from 'antd'
import { useEffect, useState } from 'react'
import { apiClient } from '../../utils/api'

interface LoginDialogProps {
 open:boolean
 onClose:()=>void
 onAuthenticate:(input:{email:string;password?:string;code?:string;type:'password'|'code'})=>Promise<boolean>
 onRegister:(input:{email:string;password:string;confirm_password:string;code?:string})=>Promise<boolean>
}
type AuthMode='login'|'register'
type AuthMethod='code'|'password'

export function LoginDialog({open,onClose,onAuthenticate,onRegister}:LoginDialogProps){
 const [mode,setMode]=useState<AuthMode>('login'),[method,setMethod]=useState<AuthMethod>('password'),[countdown,setCountdown]=useState(0),[agreement,setAgreement]=useState<'user'|'privacy'|null>(null),[registerVerification,setRegisterVerification]=useState(true)
 const [form]=Form.useForm();const [msg,ctx]=message.useMessage()
 useEffect(()=>{if(!countdown)return;const timer=window.setInterval(()=>setCountdown(v=>Math.max(0,v-1)),1000);return()=>window.clearInterval(timer)},[countdown])
 useEffect(()=>{if(open){setMode('login');setMethod('password');setCountdown(0);form.resetFields();void apiClient.auth.registrationSettings().then(settings=>setRegisterVerification(settings.register_email_verification)).catch(()=>setRegisterVerification(true))}},[open,form])
 const changeMode=(nextMode:AuthMode)=>{setMode(nextMode);setMethod(nextMode==='login'?'password':'code');setCountdown(0);form.resetFields()}
 const sendCode=async()=>{
  try{
   const {email}=await form.validateFields(['email'])
   await apiClient.auth.sendCode({email,type:mode==='register'?'register':'login'})
   setCountdown(60)
   msg.success('验证码已发送，请查收邮箱')
  }catch(error){if(error instanceof Error)msg.error(error.message)}
 }
 const [submitting,setSubmitting]=useState(false)
 const submit=async(values:{email:string;password?:string;confirmPassword?:string;code?:string})=>{
  setSubmitting(true)
  try{
   if(mode==='register'){
    const success=await onRegister({email:values.email,password:values.password||'',confirm_password:values.confirmPassword||'',code:values.code||''})
    if(!success)msg.error(registerVerification?'注册失败，请检查邮箱验证码后重试':'注册失败，请检查填写信息后重试')
    return
   }
   const success=await onAuthenticate(method==='password'
    ? {email:values.email,password:values.password,type:'password'}
    : {email:values.email,code:values.code,type:'code'})
   if(!success)msg.error(method==='password'?'邮箱或密码错误':'验证码错误或已过期')
  }catch(error){msg.error(error instanceof Error?error.message:'操作失败')}finally{setSubmitting(false)}
 }
 const action=mode==='login'?'登录':'注册'
 const isRegister=mode==='register'
 return <Modal className="auth-modal" width={440} title={null} open={open} footer={null} onCancel={onClose} destroyOnHidden>{ctx}
  <div className="auth-heading"><img src="/logo/website-logo.png" alt=""/><div><h2>{mode==='login'?'登录潮汐 AI，继续创作':'注册潮汐 AI，开始创作'}</h2><p>{mode==='login'?'':'填写邮箱完成账户创建'}</p></div></div>
  {!isRegister&&<Segmented block value={method} onChange={value=>{setMethod(value as AuthMethod);form.resetFields(['code','password'])}} options={[{label:'密码登录',value:'password'},{label:'验证码登录',value:'code'}]}/>}
  <Form form={form} className="auth-form" layout="vertical" onFinish={submit} requiredMark={false}>
   <Form.Item label="邮箱" name="email" rules={[{required:true,message:'请输入邮箱地址'},{type:'email',message:'请输入有效的邮箱地址'}]}><Input size="large" prefix={<MailOutlined/>} autoComplete="email" placeholder="name@example.com"/></Form.Item>
   {((isRegister&&registerVerification)||method==='code')&&<Form.Item label="验证码" name="code" rules={[{required:true,message:'请输入验证码'},{len:6,message:'请输入 6 位验证码'}]}><Input size="large" inputMode="numeric" maxLength={6} placeholder="请输入 6 位验证码" suffix={<Button type="link" size="small" disabled={countdown>0} onClick={sendCode}>{countdown?`${countdown}s 后重发`:'获取验证码'}</Button>}/></Form.Item>}
   {(isRegister||method==='password')&&<>
    <Form.Item label="密码" name="password" rules={[{required:true,message:'请输入密码'},{min:8,message:'密码至少 8 位'}]}><Input.Password size="large" prefix={<LockOutlined/>} autoComplete={mode==='login'?'current-password':'new-password'} placeholder="至少 8 位密码"/></Form.Item>
    {isRegister&&<Form.Item label="确认密码" name="confirmPassword" dependencies={['password']} rules={[{required:true,message:'请再次输入密码'},({getFieldValue})=>({validator(_,value){return !value||getFieldValue('password')===value?Promise.resolve():Promise.reject(new Error('两次输入的密码不一致'))}})]}><Input.Password size="large" prefix={<LockOutlined/>} autoComplete="new-password" placeholder="再次输入密码"/></Form.Item>}
   </>}
   <Form.Item className="auth-consent" name="consent" valuePropName="checked" rules={[{validator:(_,checked)=>checked?Promise.resolve():Promise.reject(new Error('请先阅读并同意用户协议和隐私政策'))}]}>
    <Checkbox>我已阅读并同意 <button type="button" onClick={event=>{event.preventDefault();setAgreement('user')}}>《用户协议》</button>和<button type="button" onClick={event=>{event.preventDefault();setAgreement('privacy')}}>《隐私政策》</button></Checkbox>
   </Form.Item>
   <Button className="auth-submit" block size="large" type="primary" htmlType="submit" loading={submitting}>{action}</Button>
   <div className="auth-switch">{mode==='login'?'还没有账号？':'已有账号？'} <button type="button" onClick={()=>changeMode(mode==='login'?'register':'login')}>{mode==='login'?'立即注册':'立即登录'}</button></div>
  </Form>
  <Modal title={agreement==='user'?'用户协议':'隐私政策'} open={!!agreement} footer={<Button type="primary" onClick={()=>setAgreement(null)}>我知道了</Button>} onCancel={()=>setAgreement(null)} destroyOnHidden>
   {agreement==='user'?<div className="agreement-content"><p>欢迎使用潮汐 AI。请合法、善意地使用本服务，不得上传违法、侵权或损害他人权益的内容。</p><p>您应对账号安全及所发布、生成的内容负责；我们可为维护服务安全而处理违规内容或限制相关账号。</p><p>本服务持续迭代，功能、模型与规则可能调整，重要变更会以适当方式通知。</p></div>:<div className="agreement-content"><p>我们仅在提供登录、创作、资产管理和必要安全保障时处理您的邮箱、账号信息及您主动提交的内容。</p><p>您的信息将用于完成服务与改进体验；除法律要求或取得您的授权外，不会向无关第三方出售或公开。</p><p>您可通过账户设置或联系平台处理个人信息相关请求。</p></div>}
  </Modal>
 </Modal>
}
