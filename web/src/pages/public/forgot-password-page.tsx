import { useEffect, useState } from 'react'
import axios from 'axios'
import { Link } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'

import { forgotPassword } from '../../lib/auth'
import { extractErrorMessage } from '../../lib/error-handler'

const defaultMailCooldownSeconds = 120

const forgotPasswordSchema = z.object({
  email: z.email(),
})

type ForgotPasswordForm = z.infer<typeof forgotPasswordSchema>

export function ForgotPasswordPage() {
  const { t, i18n } = useTranslation()
  const [cooldownSeconds, setCooldownSeconds] = useState(0)
  const [serverError, setServerError] = useState('')
  const [submitSuccess, setSubmitSuccess] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordForm>({ resolver: zodResolver(forgotPasswordSchema) })

  useEffect(() => {
    if (cooldownSeconds <= 0) return
    const timer = window.setInterval(() => {
      setCooldownSeconds((prev) => (prev <= 1 ? 0 : prev - 1))
    }, 1000)
    return () => window.clearInterval(timer)
  }, [cooldownSeconds])

  const onSubmit = async (data: ForgotPasswordForm) => {
    setServerError('')
    setSubmitSuccess(false)
    try {
      const cooldownSecondsFromServer = await forgotPassword(data.email, i18n.language)
      setCooldownSeconds(cooldownSecondsFromServer ?? defaultMailCooldownSeconds)
      setSubmitSuccess(true)
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 429) {
        const retryAfterHeader = err.response.headers['retry-after']
        const retryAfterSeconds = Number.parseInt(String(retryAfterHeader ?? ''), 10)
        setServerError(t('auth.forgotPasswordRateLimited'))
        setCooldownSeconds(Number.isFinite(retryAfterSeconds) && retryAfterSeconds > 0 ? retryAfterSeconds : defaultMailCooldownSeconds)
        return
      }
      setServerError(extractErrorMessage(err, t('auth.forgotPasswordFailed')))
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <h1 className="text-xl font-semibold">{t('auth.forgotPasswordTitle')}</h1>
      <p className="mt-2 text-sm text-muted-foreground">{t('auth.forgotPasswordDesc')}</p>

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

        {submitSuccess && !serverError && (
          <p className="rounded-lg bg-success/10 px-3 py-2 text-sm text-success">
            {t('auth.forgotPasswordEmailSent')}
          </p>
        )}
        {serverError && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {serverError}
          </p>
        )}

        <button
          type="submit"
          disabled={isSubmitting || cooldownSeconds > 0}
          className="w-full rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50 shadow-sm cursor-pointer disabled:cursor-not-allowed"
        >
          {isSubmitting
            ? t('auth.sendingResetEmail')
            : cooldownSeconds > 0
              ? t('auth.sendResetEmailCooldown', { seconds: cooldownSeconds })
              : t('auth.sendResetEmail')}
        </button>
      </form>

      <Link to="/login" className="mt-4 inline-flex text-sm font-medium text-primary transition-colors hover:underline">
        {t('auth.backToLogin')}
      </Link>
    </section>
  )
}
