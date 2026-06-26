import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { uploadImageAuth } from '@/lib/console-api'
import { fieldInputCls, fieldTextareaCls, useAdminSettingsForm } from './form'
import {
  SettingField,
  SettingsPageLayout,
} from './shared'

export function AdminSiteSettingsPage() {
  const { t } = useTranslation()
  const state = useAdminSettingsForm()
  const { register, handleSubmit, setValue, watch } = state.form
  const [uploading, setUploading] = useState(false)
  const [uploadMsg, setUploadMsg] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const faviconURL = watch('favicon_url')

  const onSubmit = handleSubmit((form) => state.saveSettings({
    app_name: form.app_name,
    app_url: form.app_url,
    site_description: form.site_description,
    favicon_url: form.favicon_url,
    footer_text_1: form.footer_text_1,
    footer_link_1: form.footer_link_1,
    footer_text_2: form.footer_text_2,
    footer_link_2: form.footer_link_2,
  }))

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
      </div>

    </SettingsPageLayout>
  )
}
