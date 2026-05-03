import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { createSetupAdmin, saveTokens } from '../../lib/auth'
import { extractErrorMessage } from '../../lib/error-handler'

const setupSchema = z.object({
  name: z.string().min(1),
  email: z.email(),
  password: z.string().min(8),
})
type SetupForm = z.infer<typeof setupSchema>

export function SetupPage() {
  const { t } = useTranslation()
  const [serverError, setServerError] = useState('')

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SetupForm>({ resolver: zodResolver(setupSchema) })

  const onSubmit = async (data: SetupForm) => {
    setServerError('')
    try {
      const tokens = await createSetupAdmin(data.email, data.password, data.name)
      saveTokens(tokens)
      toast.success(t('setup.success'))
      window.location.assign('/console/admin')
    } catch (err: unknown) {
      setServerError(extractErrorMessage(err, t('setup.failed')))
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-xl border border-border bg-card p-6 text-card-foreground">
      <p className="text-sm font-medium text-primary">{t('setup.eyebrow')}</p>
      <h1 className="mt-2 text-xl font-semibold">{t('setup.title')}</h1>
      <p className="mt-2 text-sm text-muted-foreground">{t('setup.description')}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
        <div>
          <label htmlFor="setup-name" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.name')}
          </label>
          <input
            id="setup-name"
            type="text"
            autoComplete="name"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('name')}
          />
          {errors.name && <p className="mt-1 text-xs text-destructive">{t('auth.required')}</p>}
        </div>

        <div>
          <label htmlFor="setup-email" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.email')}
          </label>
          <input
            id="setup-email"
            type="email"
            autoComplete="email"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('email')}
          />
          {errors.email && <p className="mt-1 text-xs text-destructive">{t('auth.invalidEmail')}</p>}
        </div>

        <div>
          <label htmlFor="setup-password" className="mb-1 block text-sm font-medium text-foreground">
            {t('auth.password')}
          </label>
          <input
            id="setup-password"
            type="password"
            autoComplete="new-password"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20"
            {...register('password')}
          />
          {errors.password && <p className="mt-1 text-xs text-destructive">{t('auth.passwordMin')}</p>}
        </div>

        {serverError && (
          <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {serverError}
          </p>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isSubmitting ? t('setup.submitting') : t('setup.submit')}
        </button>
      </form>
    </section>
  )
}
