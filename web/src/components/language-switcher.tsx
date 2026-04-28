import { useTranslation } from 'react-i18next'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export function LanguageSwitcher() {
  const { i18n, t } = useTranslation()

  return (
    <label className="flex items-center gap-2 text-sm text-muted-foreground">
      <span>{t('nav.language', { defaultValue: 'Language' })}</span>
      <Select 
        value={i18n.language} 
        onValueChange={(val) => val !== null && void i18n.changeLanguage(val as string)}
        items={{
          'zh-CN': '中文',
          'en-US': 'English'
        }}
      >
        <SelectTrigger className="w-[100px] h-8 bg-background border-input hover:bg-accent transition-colors">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="zh-CN">中文</SelectItem>
          <SelectItem value="en-US">English</SelectItem>
        </SelectContent>
      </Select>
    </label>
  )
}
