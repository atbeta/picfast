import { InfoIcon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  heading?: string
  title: string
  description?: string
  variant?: 'info' | 'warning' | 'destructive'
  confirmLabel?: string
  cancelLabel?: string
  destructive?: boolean
  onConfirm: () => void
  loading?: boolean
}

const variantConfig = {
  info: {
    Icon: InfoIcon,
    iconClass: 'text-info',
    buttonVariant: 'default' as const,
  },
  warning: {
    Icon: TriangleAlertIcon,
    iconClass: 'text-warning',
    buttonVariant: 'default' as const,
  },
  destructive: {
    Icon: OctagonXIcon,
    iconClass: 'text-destructive',
    buttonVariant: 'destructive' as const,
  },
}

export function ConfirmDialog({
  open,
  onOpenChange,
  heading,
  title,
  description,
  variant: variantProp,
  confirmLabel,
  cancelLabel,
  destructive,
  onConfirm,
  loading,
}: ConfirmDialogProps) {
  const { t } = useTranslation()
  const variant = variantProp ?? (destructive ? 'destructive' : 'info')
  const { Icon, iconClass, buttonVariant } = variantConfig[variant]

  const finalConfirmLabel = confirmLabel ?? t('dialog.confirm', { defaultValue: '确定' })
  const finalCancelLabel = cancelLabel ?? t('dialog.cancel', { defaultValue: '取消' })
  const finalHeading = heading ?? t('dialog.operationTitle', { defaultValue: '请确认操作' })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="gap-0 overflow-hidden p-0 sm:max-w-md">
        <DialogHeader className="border-b border-border/60 bg-muted/15 px-5 py-2.5">
          <DialogTitle className="text-[15px] leading-6">{finalHeading}</DialogTitle>
        </DialogHeader>
        <div className="px-5 py-4">
          <div className="flex items-start gap-3">
            <div className={`mt-0.5 shrink-0 rounded-full p-1.5 ${variant === 'destructive' ? 'bg-destructive/10' : variant === 'warning' ? 'bg-warning/10' : 'bg-info/10'}`}>
              <Icon className={`size-4 ${iconClass}`} />
            </div>
            <div className="space-y-1.5">
              <DialogDescription className="font-medium text-foreground">{title}</DialogDescription>
              {description && <DialogDescription>{description}</DialogDescription>}
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 border-t border-border/60 px-5 py-3">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            {finalCancelLabel}
          </Button>
          <Button
            variant={buttonVariant}
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? '...' : finalConfirmLabel}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
