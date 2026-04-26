import { useTranslation } from 'react-i18next'

import { cn } from '../lib/cn'

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation()
  return (
    <label className="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-300">
      <span>{t('common.language')}</span>
      <select
        value={i18n.language}
        onChange={(e) => void i18n.changeLanguage(e.target.value)}
        className={cn(
          'rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm',
          'dark:border-zinc-700 dark:bg-zinc-900',
        )}
      >
        <option value="zh-CN">中文</option>
        <option value="en-US">English</option>
      </select>
    </label>
  )
}
