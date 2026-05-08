import { appState } from '../stores/app'
import zhCN from './zh-CN'
import enUS from './en-US'
import type { Lang } from '../stores/app'

const messages: Record<Lang, any> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

export function useTranslate() {
  function t(path: string): string {
    const keys = path.split('.')
    let obj = messages[appState.lang]
    for (const key of keys) {
      if (obj && typeof obj === 'object' && key in obj) {
        obj = obj[key]
      } else {
        return path
      }
    }
    return typeof obj === 'string' ? obj : path
  }
  return { t }
}
