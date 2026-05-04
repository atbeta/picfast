import { Link, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'

import { resetPassword } from '../../lib/auth'

const resetPasswordSchema = z.object({
  password: z.string().min(8),
  confirmPassword: z.string().min(1),
}).refine((data) => data.password === data.confirmPassword, {
  path: ['confirmPassword'],
  message: 'passwordMismatch',
})

type ResetPasswordForm = z.infer<typeof resetPasswordSchema>

export function ResetPasswordPage() {
  const { t } = useTranslation()
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting, isSubmitSuccessful },
  } = useForm<ResetPasswordForm>({ resolver: zodResolver(resetPasswordSchema) })

  const onSubmit = async (data: ResetPasswordForm) => {
    if (!token) {
      setError('root', { message: 'missingToken' })
      return
    }
    try {
      await resetPassword(token, data.password)
    } catch {
      setError('root', { message: 'resetFailed' })
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <h1 className="text-xl font-semibold">{t('auth.resetPasswordTitle')}</h1>
      <p className="mt-2 text-sm text-muted-foreground">{t('auth.resetPasswordDesc')}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <div>
          <label htmlFor="password" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.password')}
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('password')}
          />
          {errors.password && <p className="mt-1 text-xs text-destructive">{t('auth.passwordMin')}</p>}
        </div>

        <div>
          <label htmlFor="confirmPassword" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.confirmPassword')}
          </label>
          <input
            id="confirmPassword"
            type="password"
            autoComplete="new-password"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('confirmPassword')}
          />
          {errors.confirmPassword && (
            <p className="mt-1 text-xs text-destructive">
              {errors.confirmPassword.message === 'passwordMismatch'
                ? t('auth.passwordMismatch')
                : t('auth.required')}
            </p>
          )}
        </div>

        {errors.root && errors.root.message === 'missingToken' && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {t('auth.resetPasswordTokenMissing')}
          </p>
        )}
        {errors.root && errors.root.message === 'resetFailed' && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {t('auth.resetPasswordFailed')}
          </p>
        )}
        {isSubmitSuccessful && (
          <p className="rounded-lg bg-success/10 px-3 py-2 text-sm text-success">
            {t('auth.resetPasswordSuccess')}
          </p>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50 shadow-sm cursor-pointer disabled:cursor-not-allowed"
        >
          {isSubmitting ? t('auth.resettingPassword') : t('auth.resetPassword')}
        </button>
      </form>

      <Link to="/login" className="mt-4 inline-flex text-sm font-medium text-primary transition-colors hover:underline">
        {t('auth.backToLogin')}
      </Link>
    </section>
  )
}
