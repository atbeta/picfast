import type { ReactNode } from 'react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod/v4'
import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery } from '@tanstack/react-query'

import { useAuth } from '../../lib/auth-context'
import { formatFileSize } from '../../lib/upload'
import { getStrategies, type Strategy } from '../../lib/console-api'
import { extractErrorMessage } from '../../lib/error-handler'
import { storageStrategyLabel } from '../../lib/storage-strategy'
import { getSiteConfig } from '../../lib/site-config'

import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { HelpHint } from '@/components/help-hint'
import { Switch } from '@/components/ui/switch'

const profileSchema = z.object({
  name: z.string().min(1),
  password: z.string().optional(),
  defaultStrategy: z.number().optional(),
})
type ProfileForm = z.infer<typeof profileSchema>

const fieldInputCls = 'h-10 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'
const fieldDisabledCls = 'h-10 w-full rounded-lg border border-border/50 bg-muted/50 px-4 text-sm text-muted-foreground cursor-not-allowed'

function SettingField({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: ReactNode
}) {
  return (
    <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
      <div className="pt-1">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-foreground">{label}</p>
          {hint ? <HelpHint text={hint} /> : null}
        </div>
      </div>
      <div className="min-w-0">{children}</div>
    </div>
  )
}

export function SettingsPage() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()

  const [saving, setSaving] = useState(false)
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [defaultStrategy, setDefaultStrategy] = useState(0)
  const [imageQuality, setImageQuality] = useState(85)
  const [imageFormat, setImageFormat] = useState('origin')
  const [stripExif, setStripExif] = useState(true)
  const [enableWatermark, setEnableWatermark] = useState(false)
  const [watermarkText, setWatermarkText] = useState('')
  const [watermarkPosition, setWatermarkPosition] = useState('bottom-right')
  const [watermarkFontSize, setWatermarkFontSize] = useState(28)
  const [watermarkColor, setWatermarkColor] = useState('#FFFFFF')
  const [watermarkOpacity, setWatermarkOpacity] = useState(0.6)
  const [defaultCopyFormat, setDefaultCopyFormat] = useState('')
  const [copyTemplate, setCopyTemplate] = useState('')

  const { data: siteConfig } = useQuery({
    queryKey: ['site-config'],
    queryFn: getSiteConfig,
  })
  const allowUserImageProcessing = siteConfig?.allow_user_image_processing ?? true

  useEffect(() => {
    getStrategies()
      .then((list) => {
        setStrategies(list)
        const settings = (user?.settings as Record<string, unknown>) || {}
        if (settings.default_strategy) {
          setDefaultStrategy(Number(settings.default_strategy))
        }
        const processing = (settings.image_processing as Record<string, unknown>) || {}
        setImageQuality(Number(processing.image_save_quality ?? 85))
        setImageFormat(String(processing.image_save_format ?? 'origin') || 'origin')
        setStripExif(Boolean(processing.is_strip_exif ?? true))
        setEnableWatermark(Boolean(processing.is_enable_watermark ?? false))
        const watermark = (processing.watermark_configs as Record<string, unknown>) || {}
        setWatermarkText(String(watermark.text ?? ''))
        setWatermarkPosition(String(watermark.position ?? 'bottom-right'))
        setWatermarkFontSize(Number(watermark.font_size ?? 28))
        setWatermarkColor(String(watermark.color ?? '#FFFFFF'))
        setWatermarkOpacity(Number(watermark.opacity ?? 0.6))
        setDefaultCopyFormat(String(settings.default_copy_format ?? ''))
        setCopyTemplate(String(settings.copy_template ?? ''))
      })
      .catch(() => {})
  }, [user])

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: { name: user?.name ?? '', password: '' },
  })

  const onSubmit = async (data: ProfileForm) => {
    setSaving(true)
    try {
      if (allowUserImageProcessing && enableWatermark && !watermarkText.trim()) {
        toast.error(t('settings.watermarkTextRequired', { defaultValue: '启用文字水印时，请填写水印文本。' }))
        return
      }
      const payload: { name?: string; password?: string; settings?: Record<string, unknown> } = { name: data.name }
      if (data.password && data.password.length >= 8) {
        payload.password = data.password
      }
      const currentSettings = ((user?.settings as Record<string, unknown>) || {})
      const imageProcessing = allowUserImageProcessing ? {
        image_save_quality: Number.isFinite(imageQuality) ? Math.max(1, Math.min(100, imageQuality)) : 85,
        image_save_format: imageFormat === 'origin' ? 'origin' : imageFormat,
        is_strip_exif: stripExif,
        is_enable_watermark: enableWatermark,
        watermark_configs: {
          text: watermarkText.trim(),
          position: watermarkPosition,
          font_size: Number.isFinite(watermarkFontSize) ? Math.max(8, Math.min(200, watermarkFontSize)) : 28,
          color: watermarkColor,
          opacity: Number.isFinite(watermarkOpacity) ? Math.max(0, Math.min(1, watermarkOpacity)) : 0.6,
        },
      } : currentSettings.image_processing
      payload.settings = {
        ...currentSettings,
        default_strategy: defaultStrategy && defaultStrategy !== 0 ? defaultStrategy : null,
        image_processing: imageProcessing,
        default_copy_format: defaultCopyFormat.trim() || undefined,
        copy_template: copyTemplate.trim() || undefined,
      }
      await updateProfile(payload)
      toast.success(t('settings.saved'))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('settings.saveFailed')))
    } finally {
      setSaving(false)
    }
  }

  if (!user) return null

  const usagePercent = user.capacity_bytes > 0 ? Math.round((user.used_bytes / user.capacity_bytes) * 100) : 0
  const isUnlimitedCapacity = user.capacity_bytes <= 0

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">{t('page.settings.title')}</h1>

      <div className="space-y-6 pb-8">
        
        {/* Section 1: Storage usage */}
        <div className="space-y-6">
          <div>
            <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.storage', { defaultValue: '存储用量' })}</h2>
            <p className="text-sm text-muted-foreground">{t('settings.storageDesc', { defaultValue: '查看您当前账号的可用空间和已使用情况。' })}</p>
          </div>
          <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="flex items-center gap-4">
              <div className="h-2.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-primary transition-[width] duration-200 ease-out"
                  style={{ width: `${Math.min(usagePercent, 100)}%` }}
                />
              </div>
              <span className="shrink-0 text-sm font-medium text-muted-foreground">
                {formatFileSize(user.used_bytes)} / {isUnlimitedCapacity ? t('settings.unlimitedCapacity', { defaultValue: '无限制' }) : formatFileSize(user.capacity_bytes)}
              </span>
            </div>
            <p className="mt-3 text-sm text-muted-foreground/80">
              {t('settings.stats', { images: user.image_num, albums: user.album_num })}
            </p>
          </div>
        </div>

        {/* Section 2: Profile & processing form */}
        <div className="space-y-6">
          <div className="pt-4 border-t border-border/40">
            <h2 className="text-base font-semibold tracking-tight text-foreground">{t('settings.profile', { defaultValue: '个人资料' })}</h2>
            <p className="text-sm text-muted-foreground">{t('settings.profileDesc', { defaultValue: '管理您的基础信息。' })}</p>
          </div>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
              <div className="space-y-6">
                <SettingField
                  label={t('settings.email')}
                >
                  <input
                    id="email"
                    type="email"
                    value={user.email}
                    disabled
                    className={fieldDisabledCls}
                  />
                </SettingField>

                <SettingField
                  label={t('settings.name')}
                >
                  <input
                    id="name"
                    type="text"
                    placeholder={t('settings.profileNamePlaceholder', { defaultValue: '输入您的昵称' })}
                    className={fieldInputCls}
                    {...register('name')}
                  />
                  {errors.name && <p className="mt-1.5 text-xs text-destructive">{t('auth.required')}</p>}
                </SettingField>

                <SettingField
                  label={t('settings.newPassword')}
                  hint={t('settings.passwordHint', { defaultValue: '不修改请留空' })}
                >
                  <input
                    id="password"
                    type="password"
                    autoComplete="new-password"
                    placeholder={t('settings.profilePasswordPlaceholder', { defaultValue: '留空表示不修改' })}
                    className={fieldInputCls}
                    {...register('password')}
                  />
                  {errors.password && <p className="mt-1.5 text-xs text-destructive">{t('auth.passwordMin')}</p>}
                </SettingField>

                <SettingField
                  label={t('settings.defaultStrategy', { defaultValue: '默认策略' })}
                  hint={t('settings.defaultStrategyDesc', { defaultValue: '上传时默认选中的存储策略。' })}
                >
                  <Select
                    value={defaultStrategy.toString()}
                    onValueChange={(val) => val !== null && setDefaultStrategy(Number(val))}
                    items={{
                      '0': t('settings.followGroupDefault', { defaultValue: '跟随分组默认' }),
                      ...Object.fromEntries(strategies.map(s => [s.id.toString(), `${s.name} (${storageStrategyLabel(t, s.strategy_type)})`]))
                    }}
                  >
                    <SelectTrigger id="strategy" className="h-10 w-full bg-background border-input">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">{t('settings.followGroupDefault', { defaultValue: '跟随分组默认' })}</SelectItem>
                      {strategies.map((s) => (
                        <SelectItem key={s.id} value={s.id.toString()}>
                          {s.name} ({storageStrategyLabel(t, s.strategy_type)})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </SettingField>
              </div>
            </div>

            <div className="pt-4 border-t border-border/40">
              <h2 className="text-base font-semibold tracking-tight text-foreground">
                {t('settings.imageProcessing', { defaultValue: '上传处理偏好' })}
              </h2>
              <p className="text-sm text-muted-foreground">
                {allowUserImageProcessing
                  ? t('settings.imageProcessingDesc', { defaultValue: '按你的偏好处理图片，后续上传立即生效。' })
                  : t('settings.imageProcessingDisabled', { defaultValue: '管理员已禁用用户自定义处理，当前使用系统默认策略。' })}
              </p>
            </div>
            <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
              <div className="space-y-6">
                <SettingField
                  label={t('settings.imageSaveQuality', { defaultValue: '压缩质量' })}
                  hint={t('settings.imageSaveQualityHint', { defaultValue: '范围 1-100，数值越低压缩越强。' })}
                >
                  <input
                    type="number"
                    min={1}
                    max={100}
                    value={imageQuality}
                    onChange={(e) => setImageQuality(Number(e.target.value))}
                    disabled={!allowUserImageProcessing}
                    className={fieldInputCls}
                  />
                </SettingField>

                <SettingField label={t('settings.imageSaveFormat', { defaultValue: '转码格式' })}>
                  <Select
                    value={imageFormat}
                    onValueChange={(val) => setImageFormat(String(val))}
                    items={{
                      origin: t('settings.keepOriginalFormat', { defaultValue: '保持原格式' }),
                      jpeg: 'JPEG',
                      png: 'PNG',
                      webp: 'WebP',
                    }}
                    disabled={!allowUserImageProcessing}
                  >
                    <SelectTrigger className="h-10 w-full bg-background border-input">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="origin">{t('settings.keepOriginalFormat', { defaultValue: '保持原格式' })}</SelectItem>
                      <SelectItem value="jpeg">JPEG</SelectItem>
                      <SelectItem value="png">PNG</SelectItem>
                      <SelectItem value="webp">WebP</SelectItem>
                    </SelectContent>
                  </Select>
                </SettingField>

                <SettingField label={t('settings.stripExif', { defaultValue: '去除 EXIF 元数据' })}>
                  <div className="flex h-10 items-center justify-end">
                    <Switch
                      checked={stripExif}
                      onCheckedChange={setStripExif}
                      disabled={!allowUserImageProcessing}
                    />
                  </div>
                </SettingField>

                <SettingField label={t('settings.enableWatermark', { defaultValue: '启用文字水印' })}>
                  <div className="flex h-10 items-center justify-end">
                    <Switch
                      checked={enableWatermark}
                      onCheckedChange={setEnableWatermark}
                      disabled={!allowUserImageProcessing}
                    />
                  </div>
                </SettingField>

                {enableWatermark && (
                  <>
                    <SettingField label={t('settings.watermarkText', { defaultValue: '水印文本' })}>
                      <input
                        value={watermarkText}
                        onChange={(e) => setWatermarkText(e.target.value)}
                        placeholder={t('settings.watermarkTextPlaceholder', { defaultValue: '例如：© PicFast' })}
                        disabled={!allowUserImageProcessing}
                        className={fieldInputCls}
                      />
                    </SettingField>
                    <SettingField label={t('settings.watermarkPosition', { defaultValue: '水印位置' })}>
                      <Select
                        value={watermarkPosition}
                        onValueChange={(val) => setWatermarkPosition(String(val))}
                        items={{
                          'bottom-right': t('settings.watermarkBottomRight', { defaultValue: '右下角' }),
                          'bottom-left': t('settings.watermarkBottomLeft', { defaultValue: '左下角' }),
                          'top-right': t('settings.watermarkTopRight', { defaultValue: '右上角' }),
                          'top-left': t('settings.watermarkTopLeft', { defaultValue: '左上角' }),
                          center: t('settings.watermarkCenter', { defaultValue: '居中' }),
                        }}
                        disabled={!allowUserImageProcessing}
                      >
                        <SelectTrigger className="h-10 w-full bg-background border-input">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="bottom-right">{t('settings.watermarkBottomRight', { defaultValue: '右下角' })}</SelectItem>
                          <SelectItem value="bottom-left">{t('settings.watermarkBottomLeft', { defaultValue: '左下角' })}</SelectItem>
                          <SelectItem value="top-right">{t('settings.watermarkTopRight', { defaultValue: '右上角' })}</SelectItem>
                          <SelectItem value="top-left">{t('settings.watermarkTopLeft', { defaultValue: '左上角' })}</SelectItem>
                          <SelectItem value="center">{t('settings.watermarkCenter', { defaultValue: '居中' })}</SelectItem>
                        </SelectContent>
                      </Select>
                    </SettingField>
                    <SettingField label={t('settings.watermarkFontSize', { defaultValue: '水印字号' })}>
                      <input
                        type="number"
                        min={8}
                        max={200}
                        value={watermarkFontSize}
                        onChange={(e) => setWatermarkFontSize(Number(e.target.value))}
                        disabled={!allowUserImageProcessing}
                        className={fieldInputCls}
                      />
                    </SettingField>
                    <SettingField label={t('settings.watermarkOpacity', { defaultValue: '水印透明度' })}>
                      <input
                        type="number"
                        min={0}
                        max={1}
                        step={0.1}
                        value={watermarkOpacity}
                        onChange={(e) => setWatermarkOpacity(Number(e.target.value))}
                        disabled={!allowUserImageProcessing}
                        className={fieldInputCls}
                      />
                    </SettingField>
                    <SettingField label={t('settings.watermarkColor', { defaultValue: '水印颜色' })}>
                      <input
                        value={watermarkColor}
                        onChange={(e) => setWatermarkColor(e.target.value)}
                        placeholder="#FFFFFF"
                        disabled={!allowUserImageProcessing}
                        className={fieldInputCls}
                      />
                    </SettingField>
                  </>
                )}
              </div>
            </div>

            <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-6">
              <div>
                <h3 className="text-sm font-semibold text-foreground">{t('settings.copyPreferences')}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{t('settings.copyPreferencesDesc')}</p>
              </div>
              <SettingField label={t('settings.userDefaultCopyFormat')}>
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
              </SettingField>
              <SettingField label={t('settings.userCopyTemplate')}>
                <textarea
                  value={copyTemplate}
                  onChange={(e) => setCopyTemplate(e.target.value)}
                  placeholder={'![{name}]({url})'}
                  className={`${fieldInputCls} min-h-24 font-mono`}
                />
              </SettingField>
            </div>

            <div className="pt-2">
              <div className="flex justify-end mt-6">
                <Button type="submit" size="lg" disabled={saving}>
                  {saving ? t('settings.saving') : t('settings.save')}
                </Button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </section>
  )
}
