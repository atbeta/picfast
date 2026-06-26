import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Settings, X } from 'lucide-react'
import { toast } from 'sonner'

import { useAuth } from '@/lib/auth-context'
import { usePersonalization } from '@/lib/use-personalization'
import { extractErrorMessage } from '@/lib/error-handler'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

const fieldInputCls = 'h-10 w-full rounded-lg border border-border/50 bg-background/50 px-4 text-sm outline-none transition-colors duration-150 placeholder:text-muted-foreground/50 focus:border-primary focus:ring-1 focus:ring-primary/20'

function readImageProcessing(userSettings: unknown): {
  imageQuality: number
  imageFormat: string
  stripExif: boolean
  enableWatermark: boolean
  watermarkText: string
  watermarkPosition: string
  watermarkFontSize: number
  watermarkColor: string
  watermarkOpacity: number
} {
  const processing = (userSettings as Record<string, unknown> | undefined)?.image_processing as Record<string, unknown> | undefined
  const watermark = (processing?.watermark_configs as Record<string, unknown>) || {}
  return {
    imageQuality: Number(processing?.image_save_quality ?? 85),
    imageFormat: String(processing?.image_save_format ?? 'origin') || 'origin',
    stripExif: Boolean(processing?.is_strip_exif ?? true),
    enableWatermark: Boolean(processing?.is_enable_watermark ?? false),
    watermarkText: String(watermark.text ?? ''),
    watermarkPosition: String(watermark.position ?? 'bottom-right'),
    watermarkFontSize: Number(watermark.font_size ?? 28),
    watermarkColor: String(watermark.color ?? '#FFFFFF'),
    watermarkOpacity: Number(watermark.opacity ?? 0.6),
  }
}

export function ImageProcessingDialog() {
  const { t } = useTranslation()
  const { user, updateProfile } = useAuth()
  const { workflow } = usePersonalization()
  const [open, setOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState(() => readImageProcessing(user?.settings))

  if (workflow.skipImageProcessing) {
    return (
      <Button
        type="button"
        variant="outline"
        size="icon"
        title={t('settings.imageProcessing', { defaultValue: '图片处理' })}
        disabled
      >
        <Settings className="size-4" />
      </Button>
    )
  }

  const allowUserImageProcessing = workflow.allowUserImageProcessing

  const onOpenChange = (next: boolean) => {
    if (next) setForm(readImageProcessing(user?.settings))
    setOpen(next)
  }

  const onSubmit = async () => {
    if (!user) return
    if (allowUserImageProcessing && form.enableWatermark && !form.watermarkText.trim()) {
      toast.error(t('settings.watermarkTextRequired', { defaultValue: '启用文字水印时，请填写水印文本。' }))
      return
    }
    setSaving(true)
    try {
      const current = (user.settings as Record<string, unknown>) ?? {}
      const imageProcessing = allowUserImageProcessing
        ? {
            image_save_quality: Math.max(1, Math.min(100, form.imageQuality)),
            image_save_format: form.imageFormat === 'origin' ? 'origin' : form.imageFormat,
            is_strip_exif: form.stripExif,
            is_enable_watermark: form.enableWatermark,
            watermark_configs: {
              text: form.watermarkText.trim(),
              position: form.watermarkPosition,
              font_size: Math.max(8, Math.min(200, form.watermarkFontSize)),
              color: form.watermarkColor,
              opacity: Math.max(0, Math.min(1, form.watermarkOpacity)),
            },
          }
        : (current.image_processing as Record<string, unknown> | undefined) ?? null
      await updateProfile({
        settings: { ...current, image_processing: imageProcessing },
      })
      toast.success(t('settings.saved'))
      setOpen(false)
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('settings.saveFailed')))
    } finally {
      setSaving(false)
    }
  }

  const setField = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger
        render={
          <Button
            type="button"
            variant="outline"
            size="icon"
            title={t('settings.imageProcessing', { defaultValue: '图片处理' })}
          />
        }
      >
        <Settings className="size-4" />
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('settings.imageProcessing', { defaultValue: '图片处理' })}</DialogTitle>
          <p className="text-sm text-muted-foreground">
            {allowUserImageProcessing
              ? t('settings.imageProcessingDesc', { defaultValue: '按你的偏好处理图片，后续上传立即生效。' })
              : t('settings.imageProcessingDisabled', { defaultValue: '管理员已禁用用户自定义处理，当前使用系统默认策略。' })}
          </p>
        </DialogHeader>

        <div className="space-y-4">
          <Field
            label={t('settings.imageSaveQuality', { defaultValue: '压缩质量' })}
            disabled={!allowUserImageProcessing}
          >
            <input
              type="number"
              min={1}
              max={100}
              value={form.imageQuality}
              onChange={(e) => setField('imageQuality', Number(e.target.value))}
              disabled={!allowUserImageProcessing}
              className={fieldInputCls}
            />
          </Field>

          <Field
            label={t('settings.imageSaveFormat', { defaultValue: '转码格式' })}
            disabled={!allowUserImageProcessing}
          >
            <select
              value={form.imageFormat}
              onChange={(e) => setField('imageFormat', e.target.value)}
              disabled={!allowUserImageProcessing}
              className={fieldInputCls}
            >
              <option value="origin">{t('settings.keepOriginalFormat', { defaultValue: '保持原格式' })}</option>
              <option value="jpeg">JPEG</option>
              <option value="png">PNG</option>
              <option value="webp">WebP</option>
            </select>
          </Field>

          <Field label={t('settings.stripExif', { defaultValue: '去除 EXIF 元数据' })} disabled={!allowUserImageProcessing}>
            <Switch
              checked={form.stripExif}
              onCheckedChange={(v) => setField('stripExif', v)}
              disabled={!allowUserImageProcessing}
            />
          </Field>

          <Field label={t('settings.enableWatermark', { defaultValue: '启用文字水印' })} disabled={!allowUserImageProcessing}>
            <Switch
              checked={form.enableWatermark}
              onCheckedChange={(v) => setField('enableWatermark', v)}
              disabled={!allowUserImageProcessing}
            />
          </Field>

          {form.enableWatermark && (
            <div className="space-y-4 rounded-lg border border-border/60 bg-muted/30 p-4">
              <Field
                label={t('settings.watermarkText', { defaultValue: '水印文本' })}
                disabled={!allowUserImageProcessing}
              >
                <input
                  value={form.watermarkText}
                  onChange={(e) => setField('watermarkText', e.target.value)}
                  placeholder={t('settings.watermarkTextPlaceholder', { defaultValue: '例如：© PicFast' })}
                  disabled={!allowUserImageProcessing}
                  className={fieldInputCls}
                />
              </Field>
              <Field
                label={t('settings.watermarkPosition', { defaultValue: '水印位置' })}
                disabled={!allowUserImageProcessing}
              >
                <select
                  value={form.watermarkPosition}
                  onChange={(e) => setField('watermarkPosition', e.target.value)}
                  disabled={!allowUserImageProcessing}
                  className={fieldInputCls}
                >
                  <option value="bottom-right">{t('settings.watermarkBottomRight', { defaultValue: '右下角' })}</option>
                  <option value="bottom-left">{t('settings.watermarkBottomLeft', { defaultValue: '左下角' })}</option>
                  <option value="top-right">{t('settings.watermarkTopRight', { defaultValue: '右上角' })}</option>
                  <option value="top-left">{t('settings.watermarkTopLeft', { defaultValue: '左上角' })}</option>
                  <option value="center">{t('settings.watermarkCenter', { defaultValue: '居中' })}</option>
                </select>
              </Field>
              <div className="grid grid-cols-3 gap-3">
                <Field
                  label={t('settings.watermarkFontSize', { defaultValue: '字号' })}
                  disabled={!allowUserImageProcessing}
                >
                  <input
                    type="number"
                    min={8}
                    max={200}
                    value={form.watermarkFontSize}
                    onChange={(e) => setField('watermarkFontSize', Number(e.target.value))}
                    disabled={!allowUserImageProcessing}
                    className={fieldInputCls}
                  />
                </Field>
                <Field
                  label={t('settings.watermarkOpacity', { defaultValue: '透明度' })}
                  disabled={!allowUserImageProcessing}
                >
                  <input
                    type="number"
                    min={0}
                    max={1}
                    step={0.1}
                    value={form.watermarkOpacity}
                    onChange={(e) => setField('watermarkOpacity', Number(e.target.value))}
                    disabled={!allowUserImageProcessing}
                    className={fieldInputCls}
                  />
                </Field>
                <Field
                  label={t('settings.watermarkColor', { defaultValue: '颜色' })}
                  disabled={!allowUserImageProcessing}
                >
                  <input
                    value={form.watermarkColor}
                    onChange={(e) => setField('watermarkColor', e.target.value)}
                    placeholder="#FFFFFF"
                    disabled={!allowUserImageProcessing}
                    className={fieldInputCls}
                  />
                </Field>
              </div>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
          <Button
            type="button"
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={saving}
          >
            <X className="size-4" />
            {t('common.cancel', { defaultValue: '取消' })}
          </Button>
          <Button type="button" onClick={onSubmit} disabled={saving}>
            {saving ? t('settings.saving') : t('settings.save')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function Field({
  label,
  children,
  disabled,
}: {
  label: string
  children: React.ReactNode
  disabled?: boolean
}) {
  return (
    <div className="grid grid-cols-[120px_1fr] items-center gap-3">
      <label className={['text-sm font-medium', disabled ? 'text-muted-foreground' : 'text-foreground'].join(' ')}>
        {label}
      </label>
      <div className="min-w-0">{children}</div>
    </div>
  )
}
