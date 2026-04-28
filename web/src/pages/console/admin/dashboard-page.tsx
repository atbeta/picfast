import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'

import { listAdminUsers, listAdminImages, listAdminGroups, listAdminStrategies } from '../../../lib/admin-api'

export function AdminDashboardPage() {
  const { t } = useTranslation()

  const { data: usersData } = useQuery({ queryKey: ['admin-users-stats'], queryFn: () => listAdminUsers({ page: 1, page_size: 1 }) })
  const { data: imagesData } = useQuery({ queryKey: ['admin-images-stats'], queryFn: () => listAdminImages({ page: 1, page_size: 1 }) })
  const { data: groups } = useQuery({ queryKey: ['admin-groups-stats'], queryFn: listAdminGroups })
  const { data: strategies } = useQuery({ queryKey: ['admin-strategies-stats'], queryFn: listAdminStrategies })

  const stats = [
    { label: t('admin.statUsers', { defaultValue: '用户数' }), value: usersData?.total ?? 0 },
    { label: t('admin.statImages', { defaultValue: '图片数' }), value: imagesData?.total ?? 0 },
    { label: t('admin.statGroups', { defaultValue: '分组数' }), value: Array.isArray(groups) ? groups.length : 0 },
    { label: t('admin.statStrategies', { defaultValue: '存储策略' }), value: Array.isArray(strategies) ? strategies.length : 0 },
  ]

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">{t('admin.dashboardTitle', { defaultValue: '概览' })}</h1>
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map((s) => (
          <div key={s.label} className="group overflow-hidden rounded-xl border border-border/50 bg-card p-6 shadow-sm transition-all hover:shadow-md hover:-translate-y-1 hover:border-primary/30">
            <div className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">{s.label}</div>
            <div className="text-3xl font-bold tracking-tight text-foreground">{s.value}</div>
            <div className="mt-4 h-1 w-12 rounded-full bg-primary/20 transition-all group-hover:w-full group-hover:bg-primary" />
          </div>
        ))}
      </div>
    </section>
  )
}
