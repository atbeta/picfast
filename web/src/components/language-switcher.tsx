import { useTranslation } from 'react-i18next'
import { Select, SelectContent, SelectItem, SelectTrigger } from '@/components/ui/select'
import { Languages } from 'lucide-react'
import { useAuth } from '@/lib/auth-context'

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const currentLabel = i18n.language === 'zh-CN' ? '中' : 'EN'

  const handleChange = (val: unknown) => {
    const lang = String(val ?? '')
    if (!lang) return
    void i18n.changeLanguage(lang)
    if (user) {
      const current = (user.settings as Record<string, unknown>) ?? {}
      updateProfile({
        settings: { ...current, language: val },
      }).catch(() => {})
    }
  }

  return (
    <Select
      value={i18n.language}
      onValueChange={handleChange}
      items={{
        'zh-CN': '中文',
        'en-US': 'English'
      }}
    >
      <SelectTrigger
        className="h-8 rounded-lg bg-transparent border-none shadow-none hover:bg-accent/50 focus:ring-0 px-2 flex items-center justify-center gap-1.5 cursor-pointer transition-colors"
        title={t('nav.language', { defaultValue: 'Language' })}
      >
        <Languages className="size-4 text-foreground/80" />
        <span className="hidden text-[11px] font-semibold uppercase tracking-wide text-muted-foreground sm:inline-flex">
          {currentLabel}
        </span>
        <span className="sr-only">{t('nav.language', { defaultValue: 'Language' })}</span>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="zh-CN">中文</SelectItem>
        <SelectItem value="en-US">English</SelectItem>
      </SelectContent>
    </Select>
  )
}
