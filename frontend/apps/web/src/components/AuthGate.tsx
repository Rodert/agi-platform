import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useApp } from '../store/AppStore'

export function AuthGate({children}:{children:ReactNode}){
 const {user,authReady}=useApp()
 if(!authReady)return null
 return user?children:<Navigate to="/?login=1" replace/>
}
