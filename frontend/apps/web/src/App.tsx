import { Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/AppLayout'
import { ExplorePage } from './features/inspiration'
import { CreatePage } from './pages/CreatePage'
import { VideoPage } from './pages/VideoPage'
import { AssetsPage } from './features/assets'
import { TasksPage } from './pages/TasksPage'
import { ToolsPage } from './pages/ToolsPage'
import { ProfilePage } from './pages/ProfilePage'
import { MembershipPage } from './pages/MembershipPage'
import { NotFoundPage } from './pages/NotFoundPage'
import { AuthGate } from './components/AuthGate'

export function App(){return <Routes><Route element={<AppLayout/>}><Route index element={<ExplorePage/>}/><Route path="create" element={<CreatePage/>}/><Route path="video" element={<VideoPage/>}/><Route path="assets" element={<AuthGate><AssetsPage/></AuthGate>}/><Route path="tasks" element={<AuthGate><TasksPage/></AuthGate>}/><Route path="tools" element={<ToolsPage/>}/><Route path="profile" element={<AuthGate><ProfilePage/></AuthGate>}/><Route path="membership" element={<AuthGate><MembershipPage/></AuthGate>}/><Route path="*" element={<NotFoundPage/>}/></Route></Routes>}
