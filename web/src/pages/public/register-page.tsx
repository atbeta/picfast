import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function RegisterPage() {
  const { t } = useTranslation()
  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
      <h1 className="text-xl font-semibold">{t('page.register.title')}</h1>
      <p className="mt-2 text-sm text-zinc-500">Form placeholder, will wire register API next.</p>
      <Link to="/login" className="mt-4 inline-block text-sm text-blue-600 hover:underline">
        {t('nav.login')}
      </Link>
    </section>
  )
}
