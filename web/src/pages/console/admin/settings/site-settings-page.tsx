import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Controller, useWatch } from 'react-hook-form'

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { uploadImageAuth } from '@/lib/console-api'
import {
  analyticsPayload,
  fieldInputCls,
  fieldTextareaCls,
  themePayload,
  useAdminSettingsForm,
} from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminSiteSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { register, control, handleSubmit, setValue, watch } = state.form
  const [uploading, setUploading] = useState(false)
  const [uploadMsg, setUploadMsg] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const faviconURL = watch('favicon_url')
  const analyticsProvider = useWatch({ control, name: 'analytics_provider' }) || ''

  const onSubmit = handleSubmit((form) =>
    state.saveSettings({
      app_name: form.app_name,
      app_url: form.app_url,
      site_description: form.site_description,
      favicon_url: form.favicon_url,
      footer_text_1: form.footer_text_1,
      footer_link_1: form.footer_link_1,
      footer_text_2: form.footer_text_2,
      footer_link_2: form.footer_link_2,
      show_powered_by: form.show_powered_by,
      guest_upload_notice_title: form.guest_upload_notice_title,
      guest_upload_notice_subtitle: form.guest_upload_notice_subtitle,
      show_login_link: form.show_login_link,
      theme_config: themePayload(form),
      analytics_provider: form.analytics_provider,
      analytics_config: analyticsPayload(form),
    }),
  )

  const onPickFavicon = async (files: FileList | null) => {
    const file = files?.[0]
    if (!file) return
    setUploadMsg('')
    setUploading(true)
    try {
      const result = await uploadImageAuth(file, { permission: 1 })
      setValue('favicon_url', result.links.url, { shouldDirty: true })
      setUploadMsg(t('admin.faviconUploadSuccess'))
    } catch {
      setUploadMsg(t('admin.faviconUploadFailed'))
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  return (
    <SettingsPageLayout
      title={t('admin.siteSettingsTitle')}
      description={t('admin.siteSettingsDesc')}
      state={state}
      onSubmit={onSubmit}
    >
      <SettingField label={t('admin.appName')}>
        <input {...register('app_name')} className={fieldInputCls} />
      </SettingField>

      <SettingField label={t('admin.appUrl')} hint={t('admin.appUrlDesc')}>
        <input {...register('app_url')} placeholder="https://your-domain.com" className={fieldInputCls} />
      </SettingField>

      <SettingField label={t('admin.siteDescription')} hint={t('admin.siteDescriptionDesc')}>
        <textarea
          {...register('site_description')}
          placeholder={t('admin.siteDescriptionPlaceholder')}
          className={fieldTextareaCls}
        />
      </SettingField>

      <SettingField label={t('admin.faviconUrl')} hint={t('admin.faviconUrlDesc')}>
        <div className="space-y-3">
          <input {...register('favicon_url')} placeholder="https://your-domain.com/favicon.ico" className={fieldInputCls} />
          <div className="flex flex-wrap items-center gap-3">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              className="rounded-lg border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-muted disabled:opacity-50"
            >
              {uploading ? t('admin.faviconUploading') : t('admin.faviconUpload')}
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".ico,.png,.svg,image/x-icon,image/png,image/svg+xml"
              className="hidden"
              onChange={(e) => {
                void onPickFavicon(e.target.files)
              }}
            />
            {uploadMsg ? <p className="text-sm text-muted-foreground">{uploadMsg}</p> : null}
          </div>
          {faviconURL ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <img src={faviconURL} alt="favicon preview" className="h-5 w-5 rounded-sm border border-border object-contain" />
              <span>{t('admin.faviconPreview')}</span>
            </div>
          ) : null}
        </div>
      </SettingField>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionFooterInfo')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionFooterInfoDesc')}</p>
        </div>

        <div className="flex flex-col gap-4">
          <div className="space-y-4 rounded-xl border border-border/60 bg-muted/20 p-4 md:p-5">
            <SettingField label={t('admin.footerFieldText')} hint={t('admin.footerFieldTextHint')}>
              <input {...register('footer_text_1')} className={fieldInputCls} />
            </SettingField>
            <SettingField label={t('admin.footerFieldLink')} hint={t('admin.footerFieldLinkHint')}>
              <input {...register('footer_link_1')} placeholder="https://" className={fieldInputCls} />
            </SettingField>
          </div>

          <div className="space-y-4 rounded-xl border border-border/60 bg-muted/20 p-4 md:p-5">
            <SettingField label={t('admin.footerFieldText')} hint={t('admin.footerFieldTextHint')}>
              <input {...register('footer_text_2')} className={fieldInputCls} />
            </SettingField>
            <SettingField label={t('admin.footerFieldLink')} hint={t('admin.footerFieldLinkHint')}>
              <input {...register('footer_link_2')} placeholder="https://" className={fieldInputCls} />
            </SettingField>
          </div>
        </div>
        <div className="mt-4 rounded-xl border border-border/60 bg-muted/20 p-4 md:p-5">
          <Controller
            name="show_powered_by"
            control={control}
            render={({ field }) => (
              <div className="flex items-center justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-sm font-medium text-foreground">{t('admin.showPoweredBy')}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.showPoweredByDesc')}</p>
                </div>
                <Switch
                  checked={!!field.value}
                  onCheckedChange={field.onChange}
                  id="show_powered_by"
                />
              </div>
            )}
          />
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionGuestUploadNotice')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionGuestUploadNoticeDesc')}</p>
        </div>
        <div className="space-y-4 rounded-xl border border-border/60 bg-muted/20 p-4 md:p-5">
          <SettingField label={t('admin.guestUploadNoticeTitle')} hint={t('admin.guestUploadNoticeTitleHint')}>
            <input {...register('guest_upload_notice_title')} className={fieldInputCls} />
          </SettingField>
          <SettingField label={t('admin.guestUploadNoticeSubtitle')} hint={t('admin.guestUploadNoticeSubtitleHint')}>
            <textarea {...register('guest_upload_notice_subtitle')} className={fieldTextareaCls} />
          </SettingField>
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.sectionHeaderLogin')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.sectionHeaderLoginDesc')}</p>
        </div>
        <div className="rounded-xl border border-border/60 bg-muted/20 p-4 md:p-5">
          <Controller
            name="show_login_link"
            control={control}
            render={({ field }) => (
              <div className="flex items-center justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-sm font-medium text-foreground">{t('admin.showLoginLink')}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.showLoginLinkDesc')}</p>
                </div>
                <Switch
                  checked={!!field.value}
                  onCheckedChange={field.onChange}
                  id="show_login_link"
                />
              </div>
            )}
          />
        </div>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.themeCustomCss')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.themeCustomCssHint')}</p>
        </div>
        <textarea
          id="theme_custom_css"
          {...register('theme_custom_css')}
          spellCheck={false}
          className={`${fieldTextareaCls} min-h-40 font-mono`}
        />
        <p className="mt-2 text-xs leading-5 text-muted-foreground">{t('admin.themeCustomCssScope')}</p>
      </div>

      <div className="border-t border-border/40 pt-6">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-foreground">{t('admin.analyticsSettingsTitle')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('admin.analyticsSettingsDesc')}</p>
        </div>

        <SettingField label={t('admin.analyticsProvider')}>
          <Controller
            name="analytics_provider"
            control={control}
            render={({ field }) => (
              <Select
                value={field.value || ''}
                onValueChange={(val) => field.onChange(String(val))}
                items={{
                  '': t('admin.analyticsDisabled'),
                  plausible: 'Plausible',
                  umami: 'Umami',
                  ga4: 'Google Analytics 4',
                  baidu: t('admin.analyticsBaidu'),
                  custom: t('admin.analyticsCustom'),
                }}
              >
                <SelectTrigger className="h-11 w-full bg-background border-input">
                  <SelectValue placeholder={t('admin.analyticsDisabled')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">{t('admin.analyticsDisabled')}</SelectItem>
                  <SelectItem value="plausible">Plausible</SelectItem>
                  <SelectItem value="umami">Umami</SelectItem>
                  <SelectItem value="ga4">Google Analytics 4</SelectItem>
                  <SelectItem value="baidu">{t('admin.analyticsBaidu')}</SelectItem>
                  <SelectItem value="custom">{t('admin.analyticsCustom')}</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </SettingField>

        {analyticsProvider === 'plausible' && (
          <>
            <SettingField label={t('admin.analyticsDomain')}>
              <input {...register('analytics_domain')} placeholder="img.example.com" className={fieldInputCls} />
            </SettingField>
            <SettingField label={t('admin.analyticsScriptUrl')}>
              <input {...register('analytics_script_url')} placeholder="https://plausible.io/js/script.js" className={fieldInputCls} />
            </SettingField>
          </>
        )}

        {analyticsProvider === 'umami' && (
          <>
            <SettingField label={t('admin.analyticsScriptUrl')}>
              <input {...register('analytics_script_url')} placeholder="https://analytics.example.com/script.js" className={fieldInputCls} />
            </SettingField>
            <SettingField label={t('admin.analyticsWebsiteId')}>
              <input {...register('analytics_website_id')} className={fieldInputCls} />
            </SettingField>
          </>
        )}

        {analyticsProvider === 'ga4' && (
          <SettingField label={t('admin.analyticsMeasurementId')}>
            <input {...register('analytics_measurement_id')} placeholder="G-XXXXXXXXXX" className={fieldInputCls} />
          </SettingField>
        )}

        {analyticsProvider === 'baidu' && (
          <SettingField label={t('admin.analyticsSiteId')}>
            <input {...register('analytics_site_id')} className={fieldInputCls} />
          </SettingField>
        )}

        {analyticsProvider === 'custom' && (
          <SettingField label={t('admin.analyticsCustomScript')} hint={t('admin.analyticsCustomScriptDesc')}>
            <textarea {...register('analytics_custom_script')} className={fieldTextareaCls} />
          </SettingField>
        )}

        {analyticsProvider && (
          <details className="space-y-4 rounded-lg border border-border/60 p-4">
            <summary className="cursor-pointer text-sm font-medium text-muted-foreground">
              {t('admin.analyticsAdvanced', { defaultValue: 'CSP 高级配置' })}
            </summary>
            <p className="text-xs text-muted-foreground">
              {t('admin.analyticsAdvancedDesc', { defaultValue: '仅在默认 CSP 规则不满足时填写，多个域名用逗号分隔。' })}
            </p>
            <SettingField label="connect-src">
              <input
                {...register('analytics_connect_src')}
                placeholder="https://custom-api.example.com"
                className={fieldInputCls}
              />
            </SettingField>
            <SettingField label="script-src">
              <input
                {...register('analytics_script_src')}
                placeholder="https://cdn.example.com"
                className={fieldInputCls}
              />
            </SettingField>
          </details>
        )}
      </div>
    </SettingsPageLayout>
  )
}
