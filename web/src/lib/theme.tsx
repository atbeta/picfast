import { ThemeProvider as NextThemesProvider, useTheme as useNextTheme } from 'next-themes'

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider attribute="class" defaultTheme="system" enableSystem>
      {children}
    </NextThemesProvider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export { useNextTheme as useTheme }
