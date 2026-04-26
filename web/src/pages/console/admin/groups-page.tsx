import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import {
  createAdminGroup,
  deleteAdminGroup,
  listAdminGroups,
  listAdminStrategies,
  setAdminGroupStrategies,
  updateAdminGroup,
  type AdminGroup,
} from '../../../lib/admin-api'

interface GroupForm {
  name: string
  is_default: boolean
  max_size: number
  extensions: string
  limit_per_day: number
  limit_per_month: number
  strategy_ids: number[]
}

function emptyForm(): GroupForm {
  return {
    name: '',
    is_default: false,
    max_size: 5,
    extensions: 'jpg,jpeg,png,gif,webp,bmp,svg',
    limit_per_day: 300,
    limit_per_month: 9999,
    strategy_ids: [],
  }
}

function groupToForm(g: AdminGroup): GroupForm {
  const c = g.configs || {}
  return {
    name: g.name,
    is_default: g.is_default,
    max_size: Math.round(((c.maximum_file_size as number) || 5242880) / 1048576),
    extensions: ((c.accepted_extensions as string[]) || []).join(','),
    limit_per_day: (c.limit_per_day as number) || 300,
    limit_per_month: (c.limit_per_month as number) || 9999,
    strategy_ids: (g.strategy_ids || []).map(Number),
  }
}

function formToConfigs(form: GroupForm) {
  return {
    maximum_file_size: form.max_size * 1048576,
    accepted_extensions: form.extensions.split(',').map((s) => s.trim()).filter(Boolean),
    limit_per_day: form.limit_per_day,
    limit_per_month: form.limit_per_month,
  }
}

function formatSize(bytes: number): string {
  if (!bytes) return '-'
  const mb = bytes / 1048576
  return mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb.toFixed(1)} MB`
}

export function AdminGroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const { data: groups, isLoading, error } = useQuery({
    queryKey: ['admin-groups'],
    queryFn: listAdminGroups,
  })

  const { data: allStrategies = [] } = useQuery({
    queryKey: ['admin-strategies'],
    queryFn: listAdminStrategies,
  })

  // Modal state
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<AdminGroup | null>(null)
  const [form, setForm] = useState<GroupForm>(emptyForm())
  const [saving, setSaving] = useState(false)

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm())
    setShowModal(true)
  }

  const openEdit = (g: AdminGroup) => {
    setEditing(g)
    setForm(groupToForm(g))
    setShowModal(true)
  }

  const handleSave = useCallback(async () => {
    if (!form.name.trim()) return
    setSaving(true)
    try {
      const configs = formToConfigs(form)
      if (editing) {
        await updateAdminGroup(editing.id, {
          name: form.name.trim(),
          is_default: form.is_default,
          is_guest: editing.is_guest,
          configs,
        })
        await setAdminGroupStrategies(editing.id, form.strategy_ids)
      } else {
        await createAdminGroup({
          name: form.name.trim(),
          is_default: form.is_default,
          configs,
        })
      }
      setShowModal(false)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.saveFailed'))
    } finally {
      setSaving(false)
    }
  }, [form, editing, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const onDelete = useCallback(async (id: number) => {
    if (!window.confirm(t('admin.confirmDeleteGroup'))) return
    setDeleting(id)
    try {
      await deleteAdminGroup(id)
      await qc.invalidateQueries({ queryKey: ['admin-groups'] })
    } catch (err: unknown) {
      alert((err as { response?: { data?: { message?: string } } })?.response?.data?.message || t('admin.deleteFailed'))
    } finally {
      setDeleting(null)
    }
  }, [qc, t])

  const update = <K extends keyof GroupForm>(key: K, value: GroupForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const toggleStrategy = (id: number) => {
    setForm((prev) => ({
      ...prev,
      strategy_ids: prev.strategy_ids.includes(id)
        ? prev.strategy_ids.filter((x) => x !== id)
        : [...prev.strategy_ids, id],
    }))
  }

  const inputCls = 'w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900'

  const getStrategyName = (id: number) => {
    const s = allStrategies.find((x) => x.id === id)
    return s ? s.name : String(id)
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('admin.groupsTitle')}</h1>
        <button type="button" onClick={openCreate} className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300">
          {t('admin.create')}
        </button>
      </div>

      {/* Modal overlay */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={() => setShowModal(false)}>
          <div className="w-full max-w-[500px] rounded-xl bg-white p-6 shadow-xl dark:bg-zinc-800 max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <h2 className="mb-4 text-lg font-semibold">
              {editing ? t('admin.edit') : t('admin.create')}
            </h2>

            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-sm font-medium">{t('admin.colName')}</label>
                <input value={form.name} onChange={(e) => update('name', e.target.value)} placeholder={t('admin.namePlaceholder')} className={inputCls} />
              </div>

              <div className="flex items-center gap-2">
                <input type="checkbox" id="isDefault" checked={form.is_default} onChange={(e) => update('is_default', e.target.checked)} className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600" />
                <label htmlFor="isDefault" className="text-sm">{t('admin.defaultGroup', { defaultValue: '默认分组' })}</label>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">{t('admin.maxFileSize', { defaultValue: '最大文件' })}</label>
                <div className="flex items-center gap-2">
                  <input type="number" min={1} value={form.max_size} onChange={(e) => update('max_size', Number(e.target.value))} className={`${inputCls} w-32`} />
                  <span className="text-sm text-zinc-500">MB</span>
                </div>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">{t('admin.extensions', { defaultValue: '允许格式' })}</label>
                <input value={form.extensions} onChange={(e) => update('extensions', e.target.value)} placeholder="jpg,png,gif,webp" className={inputCls} />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">{t('admin.limitPerDay', { defaultValue: '每日上限' })}</label>
                <input type="number" min={0} value={form.limit_per_day} onChange={(e) => update('limit_per_day', Number(e.target.value))} className={inputCls} />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium">{t('admin.limitPerMonth', { defaultValue: '每月上限' })}</label>
                <input type="number" min={0} value={form.limit_per_month} onChange={(e) => update('limit_per_month', Number(e.target.value))} className={inputCls} />
              </div>

              {/* Strategy binding */}
              <div className="border-t border-zinc-200 pt-3 dark:border-zinc-700">
                <label className="mb-2 block text-sm font-medium">{t('admin.availableStrategies', { defaultValue: '可用策略' })}</label>
                {allStrategies.length === 0 ? (
                  <p className="text-xs text-zinc-400">{t('admin.noStrategies', { defaultValue: '暂无策略，请先创建策略' })}</p>
                ) : (
                  <div className="flex flex-wrap gap-3">
                    {allStrategies.map((s) => (
                      <label key={s.id} className="flex items-center gap-1.5 text-sm">
                        <input
                          type="checkbox"
                          checked={form.strategy_ids.includes(s.id)}
                          onChange={() => toggleStrategy(s.id)}
                          className="h-4 w-4 rounded border-zinc-300 dark:border-zinc-600"
                        />
                        {s.name}
                      </label>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="mt-5 flex justify-end gap-2">
              <button type="button" onClick={() => setShowModal(false)} className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600">
                {t('admin.cancel')}
              </button>
              <button type="button" onClick={handleSave} disabled={saving || !form.name.trim()} className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900">
                {saving ? '…' : t('admin.confirmSave')}
              </button>
            </div>
          </div>
        </div>
      )}

      {isLoading && <div className="flex justify-center py-12"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}
      {error && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">{t('admin.loadFailed')}</p>}
      {groups && groups.length === 0 && <p className="py-12 text-center text-sm text-zinc-400">{t('admin.empty')}</p>}

      {groups && groups.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-200 text-left text-xs text-zinc-500 dark:border-zinc-700">
                <th className="pb-2 pr-3 font-medium">{t('admin.colName')}</th>
                <th className="pb-2 pr-3 font-medium">{t('admin.userCount', { defaultValue: '用户数' })}</th>
                <th className="pb-2 pr-3 font-medium">{t('admin.maxFileSize', { defaultValue: '最大文件' })}</th>
                <th className="pb-2 pr-3 font-medium">{t('admin.limitPerDay', { defaultValue: '每日上限' })}</th>
                <th className="pb-2 pr-3 font-medium">{t('admin.availableStrategies', { defaultValue: '策略' })}</th>
                <th className="pb-2 font-medium">{t('admin.colActions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-100 dark:divide-zinc-800">
              {groups.map((g) => (
                <tr key={g.id}>
                  <td className="py-2 pr-3 font-medium">
                    {g.name}
                    {g.is_default && <span className="ml-2 rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{t('admin.default')}</span>}
                    {g.is_guest && <span className="ml-2 rounded bg-green-100 px-1.5 py-0.5 text-xs text-green-700 dark:bg-green-900/30 dark:text-green-400">{t('admin.guest')}</span>}
                  </td>
                  <td className="py-2 pr-3 text-zinc-500">{(g as unknown as Record<string, unknown>).user_count as number ?? '-'}</td>
                  <td className="py-2 pr-3 text-zinc-500">{formatSize((g.configs?.maximum_file_size as number) || 0)}</td>
                  <td className="py-2 pr-3 text-zinc-500">{(g.configs?.limit_per_day as number) || '-'}</td>
                  <td className="py-2 pr-3">
                    <div className="flex flex-wrap gap-1">
                      {(g.strategy_ids || []).length > 0 ? (
                        g.strategy_ids.map((id) => (
                          <span key={id} className="rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                            {getStrategyName(id)}
                          </span>
                        ))
                      ) : (
                        <span className="text-xs text-zinc-400">-</span>
                      )}
                    </div>
                  </td>
                  <td className="py-2">
                    <div className="flex gap-1">
                      <button type="button" onClick={() => openEdit(g)} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">{t('admin.edit')}</button>
                      {!g.is_default && !g.is_guest && (
                        <button type="button" onClick={() => onDelete(g.id)} disabled={deleting === g.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('admin.delete')}</button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
