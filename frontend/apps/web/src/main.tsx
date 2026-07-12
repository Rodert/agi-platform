import React from 'react'
import ReactDOM from 'react-dom/client'
import { ConfigProvider, App as AntApp, theme } from 'antd'
import { BrowserRouter } from 'react-router-dom'
import { App } from './App'
import { AppProvider } from './store/AppStore'
import 'swiper/css'
import './styles.css'
import './features/creation/creation.css'
import './features/inspiration/inspiration.css'
import './features/assets/assets.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm, token: { colorPrimary: '#54d6cf', borderRadius: 8, borderRadiusLG: 12, borderRadiusSM: 6, controlHeight: 36, colorBgBase: '#090b10', colorTextBase: '#f3f5f8', fontFamily: 'Inter, PingFang SC, Microsoft YaHei, sans-serif' }, components: { Button: { primaryShadow: 'none', borderRadius: 8 }, Input: { borderRadius: 8 }, Select: { borderRadius: 8 }, Modal: { contentBg: '#131721', headerBg: '#131721', borderRadiusLG: 14 }, Drawer: { borderRadiusLG: 14 } } }}>
      <AntApp><BrowserRouter><AppProvider><App /></AppProvider></BrowserRouter></AntApp>
    </ConfigProvider>
  </React.StrictMode>,
)
