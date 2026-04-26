import { useTranslation } from 'react-i18next'

export function GuestUploadPage() {
  const { t } = useTranslation()
  return (
    <section className="rounded-xl border border-zinc-200 bg-white p-8 dark:border-zinc-800 dark:bg-zinc-900">
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">{t('page.guestUpload.title')}</h1>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">{t('page.guestUpload.subtitle')}</p>
      <div className="mt-6 rounded-lg border border-dashed border-zinc-300 p-16 text-center text-zinc-500 dark:border-zinc-700 dark:text-zinc-400">
        Drop files here (UI placeholder)
      </div>
    </section>
  )
}
