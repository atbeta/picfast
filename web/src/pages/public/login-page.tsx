import { useEffect, useState } from 'react'
import type { TFunction } from 'i18next'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useForm, useWatch } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'
import { resendVerification } from '../../lib/auth'
import { useAuth } from '../../lib/auth-context'
import { getSiteConfig } from '../../lib/site-config'

const loginSchema = z.object({
  email: z.email(),
  password: z.string().min(1),
})
type LoginForm = z.infer<typeof loginSchema>

/** Maps known backend English messages to localized strings. */
function mapLoginApiMessage(raw: string | undefined, t: TFunction): string {
  if (!raw?.trim()) return t('auth.loginFailed')
  const key = raw.trim().toLowerCase()
  const known: Record<string, string> = {
    'invalid email or password': t('auth.loginFailed'),
    'account is frozen': t('auth.accountFrozen'),
    'email verification required': t('auth.emailVerificationRequiredLogin'),
  }
  return known[key] ?? raw
}

export function LoginPage() {
  const { t } = useTranslation()
  const { login } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [serverError, setServerError] = useState('')
  const [allowRegister, setAllowRegister] = useState(false)
  const [requireVerification, setRequireVerification] = useState(false)
  const [resendState, setResendState] = useState<'idle' | 'loading' | 'success' | 'error'>('idle')

  useEffect(() => {
    getSiteConfig()
      .then((cfg) => {
        setAllowRegister(cfg.allow_registration)
        setRequireVerification(cfg.require_email_verification)
      })
      .catch(() => {})
  }, [])

  const {
    control,
    register,
    handleSubmit,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) })
  const emailValue = useWatch({ control, name: 'email' })

  useEffect(() => {
    const email = searchParams.get('email')
    if (email) {
      setValue('email', email)
    }
  }, [searchParams, setValue])

  const onSubmit = async (data: LoginForm) => {
    setServerError('')
    setResendState('idle')
    try {
      await login(data.email, data.password)
      navigate('/console', { replace: true })
    } catch (err: unknown) {
      const raw = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      setServerError(mapLoginApiMessage(raw, t))
    }
  }

  const handleResendVerification = async () => {
    if (!emailValue) return
    setResendState('loading')
    try {
      await resendVerification(emailValue)
      setResendState('success')
    } catch {
      setResendState('error')
    }
  }

  const verificationState = searchParams.get('verification')

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <h1 className="text-xl font-semibold">{t('page.login.title')}</h1>

      {verificationState === 'sent' && (
        <p className="mt-4 rounded-lg bg-success/10 px-3 py-2 text-sm text-success">
          {t('auth.verificationSent')}
        </p>
      )}
      {verificationState === 'pending' && (
        <p className="mt-4 rounded-lg bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-300">
          {t('auth.verificationPending')}
        </p>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <div>
          <label htmlFor="email" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.email')}
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('email')}
          />
          {errors.email && <p className="mt-1 text-xs text-destructive">{t('auth.invalidEmail')}</p>}
        </div>

        <div>
          <label htmlFor="password" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.password')}
          </label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('password')}
          />
          {errors.password && <p className="mt-1 text-xs text-destructive">{t('auth.required')}</p>}
        </div>

        {serverError && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {serverError}
          </p>
        )}

        {requireVerification && emailValue && (
          <div className="space-y-2 rounded-lg border border-border bg-background px-3 py-3">
            <p className="text-xs text-muted-foreground">{t('auth.verificationHelp')}</p>
            <button
              type="button"
              onClick={handleResendVerification}
              disabled={resendState === 'loading'}
              className="text-sm font-medium text-primary transition-colors hover:underline disabled:opacity-50"
            >
              {resendState === 'loading' ? t('auth.resendingVerification') : t('auth.resendVerification')}
            </button>
            {resendState === 'success' && <p className="text-xs text-success">{t('auth.verificationResent')}</p>}
            {resendState === 'error' && <p className="text-xs text-destructive">{t('auth.verificationResendFailed')}</p>}
          </div>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50 shadow-sm cursor-pointer disabled:cursor-not-allowed"
        >
          {isSubmitting ? t('auth.loggingIn') : t('auth.login')}
        </button>
      </form>

      {allowRegister && (
        <p className="mt-4 text-center text-sm text-muted-foreground">
          {t('auth.noAccount')}{' '}
          <Link to="/register" className="text-primary hover:underline transition-colors">
            {t('nav.register')}
          </Link>
        </p>
      )}
    </section>
  )
}
