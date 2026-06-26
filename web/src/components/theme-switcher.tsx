import { Moon, Sun, Monitor } from 'lucide-react'
import { useTheme } from '../lib/theme'
import { Button } from '@/components/ui/button'

const modes = ['light', 'dark', 'system'] as const

export function ThemeSwitcher() {
  const { resolvedTheme, setTheme, theme } = useTheme()

  const next = () => {
    const idx = modes.indexOf((theme ?? 'system') as typeof modes[number])
    setTheme(modes[(idx + 1) % modes.length])
  }

  const icon = resolvedTheme === 'dark'
    ? <Moon className="size-4 text-current" />
    : <Sun className="size-4 text-current" />

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      onClick={next}
      title={theme === 'system' ? 'System' : theme === 'dark' ? 'Dark' : 'Light'}
    >
      {icon}
      {theme === 'system' && (
        <Monitor className="absolute size-2.5 text-current opacity-60" style={{ bottom: 2, right: 2 }} />
      )}
      <span className="sr-only">Toggle theme</span>
    </Button>
  )
}
