import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { usePersonalization } from '@/lib/use-personalization'
import { extractErrorMessage } from '@/lib/error-handler'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const fieldInputCls = 'h-10 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'

export function OutputPanel() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const { output } = usePersonalization()
  const [saving, setSaving] = useState(false)

  const [defaultCopyFormat, setDefaultCopyFormat] = useState(
    (user?.settings as Record<string, unknown>)?.default_copy_format as string || output.format,
  )
  const [copyTemplate, setCopyTemplate] = useState(
    (user?.settings as Record<string, unknown>)?.copy_template as string || output.template,
  )

  if (!user) return null

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const current = (user.settings as Record<string, unknown>) ?? {}
      await updateProfile({
        settings: {
          ...current,
          default_copy_format: defaultCopyFormat.trim() || undefined,
          copy_template: copyTemplate.trim() || undefined,
        },
      })
      toast.success(t('settings.saved'))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('settings.saveFailed')))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.copyPreferences')}</h2>
        <p className="text-sm text-muted-foreground">{t('settings.copyPreferencesDesc')}</p>
      </div>
      <form onSubmit={onSubmit}>
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-6">
          <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
            <div className="pt-1">
              <p className="text-sm font-medium text-foreground">{t('settings.userDefaultCopyFormat')}</p>
            </div>
            <Select
              items={{
                '': t('settings.followGroupDefault'),
                markdown: t('copy.formatMarkdown'),
                url: t('copy.formatUrl'),
                html: t('copy.formatHtml'),
                bbcode: t('copy.formatBbcode'),
                thumbnail: t('copy.formatThumbnail'),
                custom: t('copy.formatCustom'),
              }}
              value={defaultCopyFormat}
              onValueChange={(value) => setDefaultCopyFormat(String(value ?? ''))}
            >
              <SelectTrigger className="h-10 w-full bg-background border-input">
                <SelectValue placeholder={t('settings.followGroupDefault')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t('settings.followGroupDefault')}</SelectItem>
                <SelectItem value="markdown">{t('copy.formatMarkdown')}</SelectItem>
                <SelectItem value="url">{t('copy.formatUrl')}</SelectItem>
                <SelectItem value="html">{t('copy.formatHtml')}</SelectItem>
                <SelectItem value="bbcode">{t('copy.formatBbcode')}</SelectItem>
                <SelectItem value="thumbnail">{t('copy.formatThumbnail')}</SelectItem>
                <SelectItem value="custom">{t('copy.formatCustom')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
            <div className="pt-1">
              <p className="text-sm font-medium text-foreground">{t('settings.userCopyTemplate')}</p>
            </div>
            <textarea
              value={copyTemplate}
              onChange={(e) => setCopyTemplate(e.target.value)}
              placeholder={'![{name}]({url})'}
              className={`${fieldInputCls} min-h-24 font-mono`}
            />
          </div>
        </div>
        <div className="flex justify-end mt-6">
          <Button type="submit" size="lg" disabled={saving}>
            {saving ? t('settings.saving') : t('settings.save')}
          </Button>
        </div>
      </form>
    </div>
  )
}
