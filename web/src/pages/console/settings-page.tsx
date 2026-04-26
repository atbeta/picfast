import { useTranslation } from 'react-i18next'

export function SettingsPage() {
  const { t } = useTranslation()
  return (
    <section>
      <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.settings.title')}</h1>
      <p className="mt-2 text-sm text-zinc-500">User settings placeholder.</p>
    </section>
  )
}
