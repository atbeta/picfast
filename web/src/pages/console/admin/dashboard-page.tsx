import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'

import { getAdminSettings, listAdminUsers, listAdminImages, listAdminGroups, listAdminStrategies } from '../../../lib/admin-api'

export function AdminDashboardPage() {
  const { t } = useTranslation()

  const { data: usersData } = useQuery({ queryKey: ['admin-users-stats'], queryFn: () => listAdminUsers({ page: 1, page_size: 1 }) })
  const { data: imagesData } = useQuery({ queryKey: ['admin-images-stats'], queryFn: () => listAdminImages({ page: 1, page_size: 1 }) })
  const { data: groups } = useQuery({ queryKey: ['admin-groups-stats'], queryFn: listAdminGroups })
  const { data: strategies } = useQuery({ queryKey: ['admin-strategies-stats'], queryFn: listAdminStrategies })
  const { data: settings } = useQuery({ queryKey: ['admin-dashboard-settings'], queryFn: getAdminSettings })

  const stats = [
    { label: t('admin.statUsers', { defaultValue: '用户数' }), value: usersData?.total ?? 0 },
    { label: t('admin.statImages', { defaultValue: '图片数' }), value: imagesData?.total ?? 0 },
    { label: t('admin.statGroups', { defaultValue: '分组数' }), value: Array.isArray(groups) ? groups.length : 0 },
    { label: t('admin.statStrategies', { defaultValue: '存储策略' }), value: Array.isArray(strategies) ? strategies.length : 0 },
  ]

  const siteModes = settings
    ? [
        {
          label: t('admin.allowGuestUpload'),
          value: settings.allow_guest_upload ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.allow_guest_upload ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.allowRegistration'),
          value: settings.allow_registration ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.allow_registration ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.requireEmailVerification'),
          value: settings.require_email_verification ? t('admin.statusEnabled', { defaultValue: '已开启' }) : t('admin.statusDisabled', { defaultValue: '已关闭' }),
          tone: settings.require_email_verification ? 'text-white border-transparent bg-success' : 'text-muted-foreground border-transparent bg-muted',
        },
        {
          label: t('admin.moderationMode'),
          value:
            settings.moderation_mode === 'manual'
              ? t('admin.modManual')
              : settings.moderation_mode === 'auto'
                ? t('admin.modAuto')
                : t('admin.modDisabled'),
          tone:
            settings.moderation_mode === 'manual'
              ? 'text-white border-transparent bg-amber-500'
              : settings.moderation_mode === 'auto'
                ? 'text-white border-transparent bg-info'
                : 'text-muted-foreground border-transparent bg-muted',
        },
      ]
    : []

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('admin.dashboardTitle', { defaultValue: '概览' })}</h1>
      </div>
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map((s) => (
          <div key={s.label} className="group overflow-hidden rounded-xl border border-border/50 bg-card p-6 shadow-sm transition-colors duration-150 hover:shadow-sm hover:border-primary/30">
            <div className="mb-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">{s.label}</div>
            <div className="text-3xl font-bold tracking-tight text-foreground">{s.value}</div>
          </div>
        ))}
      </div>

      {siteModes.length > 0 && (
        <div className="rounded-xl border border-border/50 bg-card p-6 shadow-sm">
          <div className="mb-5">
            <h2 className="text-lg font-semibold tracking-tight text-foreground">{t('admin.dashboardSiteState', { defaultValue: '站点状态' })}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t('admin.dashboardSiteStateDesc', { defaultValue: '快速确认当前注册、游客上传、邮箱验证与审核策略。' })}</p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {siteModes.map((item) => (
              <div key={item.label} className="rounded-lg border border-border/50 bg-background/50 p-4">
                <div className="text-sm font-medium text-foreground">{item.label}</div>
                <div className={`mt-3 inline-flex rounded-full border px-2.5 py-1 text-xs font-medium ${item.tone}`}>
                  {item.value}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </section>
  )
}
