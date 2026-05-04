import axios from 'axios'
import i18next from 'i18next'

function mapKnownApiMessage(message: string): string {
  const normalized = message.trim().toLowerCase()

  const fileSizeMatch = /^file size exceeds maximum \((\d+) bytes\)$/.exec(normalized)
  if (fileSizeMatch) {
    return i18next.t('errors.fileSizeExceedsMaximum', { bytes: fileSizeMatch[1] })
  }

  switch (normalized) {
    case 'invalid email or password':
      return i18next.t('auth.loginFailed')
    case 'email already registered':
      return i18next.t('auth.emailAlreadyRegistered')
    case 'account is frozen':
      return i18next.t('auth.accountFrozen')
    case 'email verification required':
      return i18next.t('auth.emailVerificationRequiredLogin')
    case 'invalid image data':
      return i18next.t('errors.invalidImageData')
    default:
      return message
  }
}

export function extractErrorMessage(error: unknown, fallback = '操作失败'): string {
  if (axios.isAxiosError(error)) {
    const message = (error.response?.data as { message?: string } | undefined)?.message
    if (message) return mapKnownApiMessage(message)
  }
  if (error instanceof Error && error.message) {
    return mapKnownApiMessage(error.message)
  }
  return fallback
}

export function logError(context: string, error: unknown) {
  console.error(`[${context}]`, error)
}
