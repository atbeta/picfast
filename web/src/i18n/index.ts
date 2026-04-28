import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import { defaultLng, resources, supportedLngs, type SupportedLng } from './resources'

const LANGUAGE_STORAGE_KEY = 'language'

function detectLanguage(): SupportedLng {
  const saved = localStorage.getItem(LANGUAGE_STORAGE_KEY) as SupportedLng | null
  if (saved && supportedLngs.includes(saved)) {
    return saved
  }
  const fromNavigator = navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US'
  return fromNavigator
}

void i18n.use(initReactI18next).init({
  resources,
  lng: detectLanguage(),
  fallbackLng: defaultLng,
  interpolation: { escapeValue: false },
})

i18n.on('languageChanged', (lng) => {
  localStorage.setItem(LANGUAGE_STORAGE_KEY, lng)
})

export default i18n
