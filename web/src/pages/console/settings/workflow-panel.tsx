import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { usePersonalization } from '@/lib/use-personalization'
import { getStrategies, type Strategy } from '@/lib/console-api'
import { extractErrorMessage } from '@/lib/error-handler'
import { storageStrategyLabel } from '@/lib/storage-strategy'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

const fieldInputCls = 'h-10 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'

export function WorkflowPanel() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const { workflow } = usePersonalization()
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
      })
      .catch(() => {})
  }, [user])

  if (!user) return null

  const allowUserImageProcessing = workflow.allowUserImageProcessing
  const skipImageProcessing = workflow.skipImageProcessing

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      if (allowUserImageProcessing && enableWatermark && !watermarkText.trim()) {
        toast.error(t('settings.watermarkTextRequired', { defaultValue: '启用文字水印时，请填写水印文本。' }))
        return
      }
      const current = (user.settings as Record<string, unknown>) ?? {}
      const imageProcessing = allowUserImageProcessing ? {
        image_save_quality: Math.max(1, Math.min(100, imageQuality)),
        image_save_format: imageFormat === 'origin' ? 'origin' : imageFormat,
        is_strip_exif: stripExif,
        is_enable_watermark: enableWatermark,
        watermark_configs: {
          text: watermarkText.trim(),
          position: watermarkPosition,
          font_size: Math.max(8, Math.min(200, watermarkFontSize)),
          color: watermarkColor,
          opacity: Math.max(0, Math.min(1, watermarkOpacity)),
        },
      } : current.image_processing
      await updateProfile({
        settings: {
          ...current,
          default_strategy: defaultStrategy && defaultStrategy !== 0 ? defaultStrategy : null,
          image_processing: imageProcessing,
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
      {!skipImageProcessing && (
      <>
        <div>
          <h2 className="text-base font-semibold tracking-tight text-foreground">
            {t('settings.imageProcessing', { defaultValue: '上传处理偏好' })}
          </h2>
          <p className="text-sm text-muted-foreground">
            {allowUserImageProcessing
              ? t('settings.imageProcessingDesc', { defaultValue: '按你的偏好处理图片，后续上传立即生效。' })
              : t('settings.imageProcessingDisabled', { defaultValue: '管理员已禁用用户自定义处理，当前使用系统默认策略。' })}
          </p>
        </div>
        <form onSubmit={onSubmit}>
          <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
            <div className="space-y-6">
              <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                <div className="pt-1">
                  <p className="text-sm font-medium text-foreground">{t('settings.defaultStrategy', { defaultValue: '默认策略' })}</p>
                </div>
                <Select
                  value={defaultStrategy.toString()}
                  onValueChange={(val) => val !== null && setDefaultStrategy(Number(val))}
                  items={{
                    '0': t('settings.followGroupDefault', { defaultValue: '跟随分组默认' }),
                    ...Object.fromEntries(strategies.map(s => [s.id.toString(), `${s.name} (${storageStrategyLabel(t, s.strategy_type)})`]))
                  }}
                >
                  <SelectTrigger className="h-10 w-full bg-background border-input">
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
              </div>

              <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                <div className="pt-1">
                  <p className="text-sm font-medium text-foreground">{t('settings.imageSaveQuality', { defaultValue: '压缩质量' })}</p>
                </div>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={imageQuality}
                  onChange={(e) => setImageQuality(Number(e.target.value))}
                  disabled={!allowUserImageProcessing}
                  className={fieldInputCls}
                />
              </div>

              <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                <div className="pt-1">
                  <p className="text-sm font-medium text-foreground">{t('settings.imageSaveFormat', { defaultValue: '转码格式' })}</p>
                </div>
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
              </div>

              <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                <div className="pt-1">
                  <p className="text-sm font-medium text-foreground">{t('settings.stripExif', { defaultValue: '去除 EXIF 元数据' })}</p>
                </div>
                <div className="flex h-10 items-center justify-end">
                  <Switch checked={stripExif} onCheckedChange={setStripExif} disabled={!allowUserImageProcessing} />
                </div>
              </div>

              <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                <div className="pt-1">
                  <p className="text-sm font-medium text-foreground">{t('settings.enableWatermark', { defaultValue: '启用文字水印' })}</p>
                </div>
                <div className="flex h-10 items-center justify-end">
                  <Switch checked={enableWatermark} onCheckedChange={setEnableWatermark} disabled={!allowUserImageProcessing} />
                </div>
              </div>

              {enableWatermark && (
                <>
                  <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                    <div className="pt-1"><p className="text-sm font-medium text-foreground">{t('settings.watermarkText', { defaultValue: '水印文本' })}</p></div>
                    <input value={watermarkText} onChange={(e) => setWatermarkText(e.target.value)} placeholder={t('settings.watermarkTextPlaceholder', { defaultValue: '例如：© PicFast' })} disabled={!allowUserImageProcessing} className={fieldInputCls} />
                  </div>
                  <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                    <div className="pt-1"><p className="text-sm font-medium text-foreground">{t('settings.watermarkPosition', { defaultValue: '水印位置' })}</p></div>
                    <Select value={watermarkPosition} onValueChange={(val) => setWatermarkPosition(String(val))} items={{ 'bottom-right': t('settings.watermarkBottomRight', { defaultValue: '右下角' }), 'bottom-left': t('settings.watermarkBottomLeft', { defaultValue: '左下角' }), 'top-right': t('settings.watermarkTopRight', { defaultValue: '右上角' }), 'top-left': t('settings.watermarkTopLeft', { defaultValue: '左上角' }), center: t('settings.watermarkCenter', { defaultValue: '居中' }) }} disabled={!allowUserImageProcessing}>
                      <SelectTrigger className="h-10 w-full bg-background border-input"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="bottom-right">{t('settings.watermarkBottomRight', { defaultValue: '右下角' })}</SelectItem>
                        <SelectItem value="bottom-left">{t('settings.watermarkBottomLeft', { defaultValue: '左下角' })}</SelectItem>
                        <SelectItem value="top-right">{t('settings.watermarkTopRight', { defaultValue: '右上角' })}</SelectItem>
                        <SelectItem value="top-left">{t('settings.watermarkTopLeft', { defaultValue: '左上角' })}</SelectItem>
                        <SelectItem value="center">{t('settings.watermarkCenter', { defaultValue: '居中' })}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                    <div className="pt-1"><p className="text-sm font-medium text-foreground">{t('settings.watermarkFontSize', { defaultValue: '水印字号' })}</p></div>
                    <input type="number" min={8} max={200} value={watermarkFontSize} onChange={(e) => setWatermarkFontSize(Number(e.target.value))} disabled={!allowUserImageProcessing} className={fieldInputCls} />
                  </div>
                  <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                    <div className="pt-1"><p className="text-sm font-medium text-foreground">{t('settings.watermarkOpacity', { defaultValue: '水印透明度' })}</p></div>
                    <input type="number" min={0} max={1} step={0.1} value={watermarkOpacity} onChange={(e) => setWatermarkOpacity(Number(e.target.value))} disabled={!allowUserImageProcessing} className={fieldInputCls} />
                  </div>
                  <div className="grid gap-3 md:grid-cols-[180px_minmax(0,1fr)] md:items-start md:gap-6">
                    <div className="pt-1"><p className="text-sm font-medium text-foreground">{t('settings.watermarkColor', { defaultValue: '水印颜色' })}</p></div>
                    <input value={watermarkColor} onChange={(e) => setWatermarkColor(e.target.value)} placeholder="#FFFFFF" disabled={!allowUserImageProcessing} className={fieldInputCls} />
                  </div>
                </>
              )}
            </div>
          </div>
          <div className="flex justify-end mt-6">
            <Button type="submit" size="lg" disabled={saving}>
              {saving ? t('settings.saving') : t('settings.save')}
            </Button>
          </div>
        </form>
      </>
      )}
    </div>
  )
}
