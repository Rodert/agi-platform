import React from 'react'
import ReactDOM from 'react-dom/client'
import { App as AntApp } from 'antd'
import { BrowserRouter } from 'react-router-dom'
import { App } from './App'
import { AppProvider } from './store/AppStore'
import { ThemeProvider } from './store/ThemeStore'
import 'swiper/css'
import './styles.css'
import './features/creation/creation.css'
import './features/inspiration/inspiration.css'
import './features/assets/assets.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeProvider><AntApp><BrowserRouter><AppProvider><App /></AppProvider></BrowserRouter></AntApp></ThemeProvider>
  </React.StrictMode>,
)
