import { Moon, Sun } from 'lucide-react'
import { useTheme } from '../lib/theme'
import { Button } from '@/components/ui/button'

export function ThemeSwitcher() {
  const { resolvedTheme, setTheme } = useTheme()

  return (
    <Button
      variant="ghost"
      size="icon-sm"
      onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
    >
      <Sun className="size-4 text-current transition-opacity duration-150 dark:opacity-0" />
      <Moon className="absolute size-4 text-current opacity-0 transition-opacity duration-150 dark:opacity-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  )
}
