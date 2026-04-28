import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'
import { useAuth } from '../../lib/auth-context'
import { getSiteConfig } from '../../lib/site-config'

const loginSchema = z.object({
  email: z.email(),
  password: z.string().min(1),
})
type LoginForm = z.infer<typeof loginSchema>

export function LoginPage() {
  const { t } = useTranslation()
  const { login } = useAuth()
  const navigate = useNavigate()
  const [serverError, setServerError] = useState('')
  const [allowRegister, setAllowRegister] = useState(false)

  useEffect(() => {
    getSiteConfig().then((cfg) => setAllowRegister(cfg.allow_registration)).catch(() => {})
  }, [])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) })

  const onSubmit = async (data: LoginForm) => {
    setServerError('')
    try {
      await login(data.email, data.password)
      navigate('/console', { replace: true })
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('auth.loginFailed')
      setServerError(msg)
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <h1 className="text-xl font-semibold">{t('page.login.title')}</h1>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <div>
          <label htmlFor="email" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.email')}
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
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
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary/20 transition-all"
            {...register('password')}
          />
          {errors.password && <p className="mt-1 text-xs text-destructive">{t('auth.required')}</p>}
        </div>

        {serverError && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {serverError}
          </p>
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
