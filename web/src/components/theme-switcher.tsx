import { useTranslation } from 'react-i18next'

import { useTheme } from '../lib/theme'
import { cn } from '../lib/cn'

export function ThemeSwitcher() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  return (
    <label className="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-300">
      <span>{t('common.theme')}</span>
      <select
        value={theme}
        onChange={(e) => setTheme(e.target.value as 'light' | 'dark' | 'system')}
        className={cn(
          'rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm',
          'dark:border-zinc-700 dark:bg-zinc-900',
        )}
      >
        <option value="light">light</option>
        <option value="dark">dark</option>
        <option value="system">system</option>
      </select>
    </label>
  )
}
