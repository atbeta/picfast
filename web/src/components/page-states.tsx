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
        'flex flex-col items-center justify-center rounded-2xl border border-dashed border-border/60 bg-muted/10 text-center',
        compact ? 'px-6 py-12' : 'px-8 py-24',
        className,
      ].join(' ')}
    >
      {icon && (
        <div className="mb-5 flex h-16 w-16 items-center justify-center rounded-full bg-muted/30 border border-border/40 shadow-sm">
          <div className="text-muted-foreground/70 scale-[1.3]">{icon}</div>
        </div>
      )}
      <p className="text-[15px] font-semibold text-foreground tracking-tight">{title}</p>
      {description && <p className="mt-2 max-w-sm text-[13px] leading-relaxed text-muted-foreground">{description}</p>}
      {action && <div className="mt-6">{action}</div>}
    </div>
  )
}
