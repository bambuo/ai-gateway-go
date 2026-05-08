import { reactive } from 'vue'

const THEME_KEY = 'ai-gateway-theme'
const LANG_KEY = 'ai-gateway-lang'
const ADMIN_KEY = 'ai-gateway-admin'

export type ThemeMode = 'light' | 'dark'
export type Lang = 'zh-CN' | 'en-US'

interface AppState {
  theme: ThemeMode
  lang: Lang
  sidebarCollapsed: boolean
  adminName: string
}

const savedTheme = (localStorage.getItem(THEME_KEY) as ThemeMode) || 'light'
const savedLang = (localStorage.getItem(LANG_KEY) as Lang) || 'zh-CN'
const savedAdmin = localStorage.getItem(ADMIN_KEY) || ''

function applyTheme(theme: ThemeMode) {
  if (theme === 'dark') {
    document.body.setAttribute('arco-theme', 'dark')
  } else {
    document.body.removeAttribute('arco-theme')
  }
}

applyTheme(savedTheme)

export const appState = reactive<AppState>({
  theme: savedTheme,
  lang: savedLang,
  sidebarCollapsed: false,
  adminName: savedAdmin,
})

export function toggleTheme() {
  const next = appState.theme === 'light' ? 'dark' : 'light'
  appState.theme = next
  localStorage.setItem(THEME_KEY, next)
  applyTheme(next)
}

export function setLang(lang: Lang) {
  appState.lang = lang
  localStorage.setItem(LANG_KEY, lang)
}

export function setAdminName(name: string) {
  appState.adminName = name
  localStorage.setItem(ADMIN_KEY, name)
}

export function clearAdmin() {
  appState.adminName = ''
  localStorage.removeItem(ADMIN_KEY)
}
