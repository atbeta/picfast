import { InfoIcon, OctagonXIcon, TriangleAlertIcon } from 'lucide-react'
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
  title,
  description,
  variant: variantProp,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive,
  onConfirm,
  loading,
}: ConfirmDialogProps) {
  const variant = variantProp ?? (destructive ? 'destructive' : 'info')
  const { Icon, iconClass, buttonVariant } = variantConfig[variant]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <div className="flex items-start gap-3">
            <div className={`mt-0.5 shrink-0 rounded-full p-1.5 ${variant === 'destructive' ? 'bg-destructive/10' : variant === 'warning' ? 'bg-warning/10' : 'bg-info/10'}`}>
              <Icon className={`size-4 ${iconClass}`} />
            </div>
            <div>
              <DialogTitle>{title}</DialogTitle>
              {description && <DialogDescription className="mt-1.5">{description}</DialogDescription>}
            </div>
          </div>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            variant={buttonVariant}
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? '...' : confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
