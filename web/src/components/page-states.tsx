import type { ReactNode } from 'react'
import { LoaderCircle } from 'lucide-react'

interface LoadingStateProps {
  label?: string
  className?: string
  compact?: boolean
}

interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description?: string
  action?: ReactNode
  className?: string
  compact?: boolean
}

export function LoadingState({
  label,
  className = '',
  compact = false,
}: LoadingStateProps) {
  return (
    <div
      className={[
        'flex flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-border/60 bg-muted/15 text-center',
        compact ? 'py-8' : 'py-14',
        className,
      ].join(' ')}
    >
      <div className="rounded-full bg-background/80 p-3 shadow-sm">
        <LoaderCircle className="size-5 animate-spin text-primary" />
      </div>
      {label && <p className="text-sm text-muted-foreground">{label}</p>}
    </div>
  )
}

export function EmptyState({
  icon,
  title,
  description,
  action,
  className = '',
  compact = false,
}: EmptyStateProps) {
  return (
    <div
      className={[
        'flex flex-col items-center justify-center rounded-xl border border-dashed border-border/60 bg-muted/15 text-center',
        compact ? 'px-5 py-8' : 'px-6 py-14',
        className,
      ].join(' ')}
    >
      {icon && <div className="mb-4 rounded-full bg-background/70 p-4 shadow-sm">{icon}</div>}
      <p className="text-sm font-semibold text-foreground">{title}</p>
      {description && <p className="mt-2 max-w-md text-sm text-muted-foreground">{description}</p>}
      {action && <div className="mt-5">{action}</div>}
    </div>
  )
}
