import { ConfigProvider, theme as antdTheme } from 'antd'
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export type ThemePreference = 'light' | 'dark' | 'system'
type ResolvedTheme = Exclude<ThemePreference, 'system'>

interface ThemeStore {
  preference: ThemePreference
  resolvedTheme: ResolvedTheme
  setPreference: (preference: ThemePreference) => void
}

const STORAGE_KEY = 'agi-theme-preference'
const ThemeContext = createContext<ThemeStore | null>(null)

const getStoredPreference = (): ThemePreference => {
  const stored = window.localStorage.getItem(STORAGE_KEY)
  return stored === 'light' || stored === 'dark' || stored === 'system' ? stored : 'system'
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>(getStoredPreference)
  const [systemTheme, setSystemTheme] = useState<ResolvedTheme>(() =>
    window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light',
  )
  const resolvedTheme = preference === 'system' ? systemTheme : preference

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const update = () => setSystemTheme(media.matches ? 'dark' : 'light')
    update()
    media.addEventListener('change', update)
    return () => media.removeEventListener('change', update)
  }, [])

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme
    document.documentElement.style.colorScheme = resolvedTheme
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', resolvedTheme === 'dark' ? '#090b10' : '#f5f7fa')
  }, [resolvedTheme])

  const value = useMemo<ThemeStore>(() => ({
    preference,
    resolvedTheme,
    setPreference: nextPreference => {
      window.localStorage.setItem(STORAGE_KEY, nextPreference)
      setPreferenceState(nextPreference)
    },
  }), [preference, resolvedTheme])

  const isDark = resolvedTheme === 'dark'
  return <ThemeContext.Provider value={value}>
    <ConfigProvider theme={{
      algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
      token: {
        colorPrimary: '#089e9b',
        borderRadius: 8,
        borderRadiusLG: 12,
        borderRadiusSM: 6,
        controlHeight: 36,
        colorBgBase: isDark ? '#090b10' : '#f5f7fa',
        colorTextBase: isDark ? '#f3f5f8' : '#1d2733',
        fontFamily: 'Inter, PingFang SC, Microsoft YaHei, sans-serif',
      },
      components: {
        Button: { primaryShadow: 'none', borderRadius: 8 },
        Input: { borderRadius: 8 },
        Select: { borderRadius: 8 },
        Modal: { contentBg: isDark ? '#131721' : '#ffffff', headerBg: isDark ? '#131721' : '#ffffff', borderRadiusLG: 14 },
        Drawer: { borderRadiusLG: 14 },
      },
    }}>
      {children}
    </ConfigProvider>
  </ThemeContext.Provider>
}

export function useTheme() {
  const value = useContext(ThemeContext)
  if (!value) throw new Error('ThemeProvider missing')
  return value
}
