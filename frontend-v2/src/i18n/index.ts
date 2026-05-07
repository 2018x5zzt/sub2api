import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import en from './locales/en'
import zh from './locales/zh'

export type LocaleCode = 'en' | 'zh'

const LOCALE_KEY = 'sub2api_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'

export function getStoredLocale(): LocaleCode | null {
  const v = localStorage.getItem(LOCALE_KEY)
  return v === 'en' || v === 'zh' ? v : null
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      zh: { translation: zh }
    },
    fallbackLng: DEFAULT_LOCALE,
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      lookupLocalStorage: LOCALE_KEY,
      caches: ['localStorage'],
      convertDetectedLanguage: (lng) => {
        const lower = lng.toLowerCase()
        if (lower.startsWith('zh')) return 'zh'
        return 'en'
      }
    }
  })

i18n.on('languageChanged', (lng) => {
  document.documentElement.setAttribute('lang', lng)
})

export async function setLocale(locale: LocaleCode): Promise<void> {
  await i18n.changeLanguage(locale)
  localStorage.setItem(LOCALE_KEY, locale)
}

export function getLocale(): LocaleCode {
  const lng = i18n.language || DEFAULT_LOCALE
  return lng.startsWith('zh') ? 'zh' : 'en'
}

export const availableLocales = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: '中文', flag: '🇨🇳' }
] as const

export default i18n
