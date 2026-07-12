import { BellOutlined, BulbOutlined, ClockCircleOutlined, FolderOpenOutlined, HomeOutlined, LogoutOutlined, MenuOutlined, SettingOutlined, TeamOutlined, ToolOutlined, UserOutlined, WalletOutlined } from '@ant-design/icons'
import { Avatar, Badge, Button, Drawer, Dropdown, message, type MenuProps } from 'antd'
import { useEffect, useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useApp } from '../store/AppStore'
import { LOGIN_REQUIRED_EVENT } from '../utils/auth'
import { LoginDialog } from '../features/auth/LoginDialog'
import { CommunityDialog } from '../features/community/CommunityDialog'
import { NotificationsDrawer } from '../features/notifications/NotificationsDrawer'

const nav=[['/','灵感',HomeOutlined],['/create','生成',BulbOutlined],['/assets','资产',FolderOpenOutlined],['/tools','工具',ToolOutlined]] as const
export function AppLayout(){
 const {user,balance,notifications,markNotificationsRead,login,logout,requireAuth}=useApp(),navigate=useNavigate(),location=useLocation()
 const [drawer,setDrawer]=useState(false),[loginOpen,setLoginOpen]=useState(false),[communityOpen,setCommunityOpen]=useState(false),[noticeOpen,setNoticeOpen]=useState(false);const [msg,ctx]=message.useMessage()
 const doLogin=async(email:string,password:string)=>{const success=await login(email,password);if(success){setLoginOpen(false);msg.success('登录成功，欢迎回来')}return success}
 useEffect(()=>{const open=()=>setLoginOpen(true);window.addEventListener(LOGIN_REQUIRED_EVENT,open);return()=>window.removeEventListener(LOGIN_REQUIRED_EVENT,open)},[])
 useEffect(()=>{if(new URLSearchParams(location.search).get('login')==='1'){setLoginOpen(true);navigate('/',{replace:true})}},[location.search,navigate])
 const openNotices=()=>{if(!requireAuth())return;setNoticeOpen(true);markNotificationsRead()}
 const openProtected=(path:string)=>{if(requireAuth())navigate(path)}
 const accountItems:MenuProps['items']=[
  {key:'profile',icon:<UserOutlined/>,label:'个人中心'},
  {key:'account',icon:<WalletOutlined/>,label:'我的账户'},
  {key:'tasks',icon:<ClockCircleOutlined/>,label:'创作任务'},
  {type:'divider'},
  {key:'settings',icon:<SettingOutlined/>,label:'账户设置'},
  {key:'logout',icon:<LogoutOutlined/>,label:'退出登录',danger:true},
 ]
 const accountAction:MenuProps['onClick']=({key})=>{if(key==='logout'){logout();navigate('/');msg.success('已退出登录')}else if(key==='account')navigate('/membership');else if(key==='tasks')navigate('/tasks');else navigate('/profile')}
 return <div className="app-shell">{ctx}
  <aside className="app-sidebar desktop-only"><NavLink to="/" className="sidebar-logo" aria-label="潮汐 AI 首页"><img src="/logo/website-logo.png" alt=""/></NavLink><nav className="sidebar-nav">{nav.map(([to,label,Icon])=><NavLink key={to} to={to} end={to==='/'} onClick={e=>{if(to==='/assets'&&!requireAuth())e.preventDefault()}} className={({isActive})=>`sidebar-nav-item ${isActive?'active':''}`}><Icon/><span>{label}</span></NavLink>)}</nav><div className="sidebar-bottom"><button className="wave-action" title="获取灵感值" onClick={()=>openProtected('/membership')}><WalletOutlined/><b>{balance}</b></button><button className="side-action" title="加入社群" onClick={()=>setCommunityOpen(true)}><TeamOutlined/></button><Badge count={notifications} size="small"><button className="side-action" title="通知" onClick={openNotices}><BellOutlined/></button></Badge>{user?<Dropdown menu={{items:accountItems,onClick:accountAction}} placement="topLeft" trigger={['click']}><button className="side-avatar"><Avatar size={32} src={user.avatar}/></button></Dropdown>:<button className="side-action side-login" title="登录" onClick={()=>setLoginOpen(true)}><UserOutlined/></button>}</div></aside>
  <header className="mobile-only fixed inset-x-0 top-0 z-30 h-16 border-b border-[#1d222d] bg-[#090b10e8] px-4 backdrop-blur-xl"><div className="flex h-full items-center justify-between"><div className="flex items-center gap-2"><Button type="text" icon={<MenuOutlined/>} onClick={()=>setDrawer(true)}/><b>潮汐 AI</b></div><div className="flex items-center gap-1"><Button icon={<WalletOutlined/>} onClick={()=>openProtected('/membership')}><strong className="text-[#54d6cf]">{balance}</strong></Button><Badge count={notifications} size="small"><Button aria-label="通知" type="text" shape="circle" icon={<BellOutlined/>} onClick={openNotices}/></Badge>{user?<Avatar size={32} src={user.avatar} onClick={()=>navigate('/profile')}/>:<Button type="primary" onClick={()=>setLoginOpen(true)}>登录</Button>}</div></div></header>
  <div className="ml-16 max-[800px]:ml-0 max-[800px]:pt-16"><Outlet/></div>
  <nav className="mobile-only fixed bottom-0 inset-x-0 z-40 h-16 border-t border-[#242a37] bg-[#0b0e15ed] backdrop-blur-xl"><div className="flex h-full items-center justify-around">{nav.map(([to,label,Icon])=><NavLink key={to} to={to} end={to==='/'} className={({isActive})=>`grid place-items-center gap-1 text-xs ${isActive?'text-[#54d6cf]':'text-[#7f889a]'}`}><Icon className="text-lg"/><span>{label}</span></NavLink>)}</div></nav>
  <Drawer title="潮汐 AI" placement="left" width={280} open={drawer} onClose={()=>setDrawer(false)}>{nav.map(([to,label,Icon])=><Button key={to} block type="text" className="!mb-2 !h-12 !justify-start" icon={<Icon/>} onClick={()=>{navigate(to);setDrawer(false)}}>{label}</Button>)}<Button block type="text" className="!h-12 !justify-start" icon={<ClockCircleOutlined/>} onClick={()=>{navigate('/tasks');setDrawer(false)}}>创作任务</Button><Button block type="text" className="!h-12 !justify-start" icon={<TeamOutlined/>} onClick={()=>{setCommunityOpen(true);setDrawer(false)}}>加入社群</Button><Button block type="text" className="!h-12 !justify-start" icon={<BellOutlined/>} onClick={()=>{openNotices();setDrawer(false)}}>通知中心</Button><Button block type="text" className="!h-12 !justify-start" icon={<UserOutlined/>} onClick={()=>{user?navigate('/profile'):setLoginOpen(true);setDrawer(false)}}>个人中心</Button><Button block type="text" className="!h-12 !justify-start" icon={<WalletOutlined/>} onClick={()=>{navigate('/membership');setDrawer(false)}}>我的账户</Button></Drawer>
  <LoginDialog open={loginOpen} onClose={()=>setLoginOpen(false)} onLogin={doLogin}/>
  <CommunityDialog open={communityOpen} onClose={()=>setCommunityOpen(false)}/>
  <NotificationsDrawer open={noticeOpen} onClose={()=>setNoticeOpen(false)}/>
 </div>
}
