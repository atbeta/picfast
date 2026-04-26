import { useTranslation } from 'react-i18next'

export function ImagesPage() {
  const { t } = useTranslation()
  return (
    <section>
      <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.images.title')}</h1>
      <p className="mt-2 text-sm text-zinc-500">Image table placeholder.</p>
    </section>
  )
}
