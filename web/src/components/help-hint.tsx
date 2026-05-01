import { CircleHelp } from 'lucide-react'

export function HelpHint({ text }: { text: string }) {
  return (
    <span className="group relative inline-flex align-middle">
      <button
        type="button"
        tabIndex={0}
        aria-label={text}
        className="inline-flex size-4 items-center justify-center rounded-full text-muted-foreground/80 transition-colors hover:text-foreground focus:text-foreground focus:outline-none"
      >
        <CircleHelp className="size-3.5" />
      </button>
      <span className="pointer-events-none absolute top-full left-0 z-30 mt-2 hidden w-56 rounded-lg border border-border/60 bg-popover px-3 py-2 text-left text-xs leading-5 text-popover-foreground shadow-lg group-hover:block group-focus-within:block md:left-full md:top-1/2 md:mt-0 md:ml-3 md:w-64 md:-translate-y-1/2">
        {text}
      </span>
    </span>
  )
}
