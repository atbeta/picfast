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
      <Sun className="size-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute size-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      <span className="sr-only">Toggle theme</span>
    </Button>
  )
}
