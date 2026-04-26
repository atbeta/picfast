import { useTranslation } from 'react-i18next'

export function ApiTokensPage() {
  const { t } = useTranslation()
  return (
    <section>
      <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.apiTokens.title')}</h1>
      <p className="mt-2 text-sm text-zinc-500">API token management placeholder.</p>
    </section>
  )
}
