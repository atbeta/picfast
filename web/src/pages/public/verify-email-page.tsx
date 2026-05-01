import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

import { verifyEmail } from '../../lib/auth'

export function VerifyEmailPage() {
  const { t } = useTranslation()
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const [status, setStatus] = useState<'pending' | 'success' | 'error'>(token ? 'pending' : 'error')

  useEffect(() => {
    if (!token) {
      return
    }

    verifyEmail(token)
      .then(() => setStatus('success'))
      .catch(() => setStatus('error'))
  }, [token])

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <h1 className="text-xl font-semibold">{t('auth.verifyEmailTitle')}</h1>

      <div className="mt-6 rounded-lg border border-border bg-background px-4 py-5">
        {status === 'pending' && <p className="text-sm text-muted-foreground">{t('auth.verifyEmailPending')}</p>}
        {status === 'success' && <p className="text-sm text-success">{t('auth.verifyEmailSuccess')}</p>}
        {status === 'error' && <p className="text-sm text-destructive">{t('auth.verifyEmailFailed')}</p>}
      </div>

      <Link to="/login" className="mt-4 inline-flex text-sm font-medium text-primary transition-colors hover:underline">
        {t('auth.backToLogin')}
      </Link>
    </section>
  )
}
