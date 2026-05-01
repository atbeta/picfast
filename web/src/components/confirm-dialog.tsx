import { InfoIcon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
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
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-start gap-3">
            <div className={`mt-0.5 shrink-0 rounded-full p-1.5 ${variant === 'destructive' ? 'bg-destructive/10' : variant === 'warning' ? 'bg-warning/10' : 'bg-info/10'}`}>
              <Icon className={`size-4 ${iconClass}`} />
            </div>
            <div className="space-y-1">
              <DialogTitle className="text-base">{finalHeading}</DialogTitle>
              <DialogDescription className="font-medium text-foreground">{title}</DialogDescription>
              {description && <DialogDescription>{description}</DialogDescription>}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter className="mt-4">
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
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
