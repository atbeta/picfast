import { useTranslation } from 'react-i18next'

export function UploadPage() {
  const { t } = useTranslation()
  return (
    <section>
      <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.upload.title')}</h1>
      <p className="mt-2 text-sm text-zinc-500">Upload workflow placeholder.</p>
    </section>
  )
}
