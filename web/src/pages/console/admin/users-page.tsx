import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Ban, Unlock, Trash2, ChevronLeft, ChevronRight, Users, Pencil } from 'lucide-react'

import { deleteAdminUser, listAdminUsers, updateAdminUser, listAdminGroups, type AdminUser } from '../../../lib/admin-api'
import { extractErrorMessage } from '../../../lib/error-handler'
import { formatFileSize } from '../../../lib/upload'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export function AdminUsersPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const pageSize = 20

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-users', page, keyword],
    queryFn: () => listAdminUsers({ page, page_size: pageSize, keyword: keyword || undefined }),
  })
  const totalPages = data ? (data.total_pages > 0 ? data.total_pages : Math.max(1, Math.ceil(data.total / pageSize))) : 1

  const { data: groups = [] } = useQuery({
    queryKey: ['admin-groups'],
    queryFn: listAdminGroups,
  })

  const [saving, setSaving] = useState<number | null>(null)
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const [editing, setEditing] = useState<AdminUser | null>(null)
  const [editName, setEditName] = useState('')
  const [editPassword, setEditPassword] = useState('')
  const [editGroup, setEditGroup] = useState<string>('none')
  const [editCapacity, setEditCapacity] = useState<number>(0)
  const [editSaving, setEditSaving] = useState(false)

  const openEdit = (u: AdminUser) => {
    setEditing(u)
    setEditName(u.name)
    setEditPassword('')
    setEditGroup(u.group_id ? u.group_id.toString() : 'none')
    setEditCapacity(Math.round(u.capacity_bytes / 1048576))
  }

  const handleEditSave = async () => {
    if (!editing) return
    setEditSaving(true)
    try {
      await updateAdminUser(editing.id, {
        name: editName.trim() || undefined,
        password: editPassword.trim() || undefined,
        group_id: editGroup === 'none' ? null : Number(editGroup),
        capacity_bytes: editCapacity * 1048576,
      })
      setEditing(null)
      await qc.invalidateQueries({ queryKey: ['admin-users'] })
      toast.success(t('admin.saveSuccess', { defaultValue: '更新成功' }))
    } catch (err: unknown) {
      toast.error(extractErrorMessage(err, t('admin.saveFailed')))
    } finally {
      setEditSaving(false)
    }
  }

  const toggleStatus = useCallback(
    async (user: { id: number; status: number }) => {
      setSaving(user.id)
      try {
        await updateAdminUser(user.id, { status: user.status === 1 ? 0 : 1 })
        await qc.invalidateQueries({ queryKey: ['admin-users'] })
      } catch (err: unknown) {
        toast.error(extractErrorMessage(err, t('admin.saveFailed')))
      } finally {
        setSaving(null)
      }
    },
    [qc, t],
  )

  const onDelete = useCallback(
    async () => {
      if (deleteTarget === null) return
      setDeleting(deleteTarget)
      try {
        await deleteAdminUser(deleteTarget)
        setDeleteTarget(null)
        await qc.invalidateQueries({ queryKey: ['admin-users'] })
      } catch (err: unknown) {
        toast.error(extractErrorMessage(err, t('admin.deleteFailed')))
      } finally {
        setDeleting(null)
      }
    },
    [deleteTarget, qc, t],
  )

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{t('admin.usersTitle')}</h1>
          <p className="text-sm text-muted-foreground">{t('admin.usersSubtitle', { defaultValue: '集中管理用户账户、状态与容量配额。' })}</p>
        </div>
        <input
          value={keyword}
          onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
          placeholder={t('admin.searchPlaceholder')}
          className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-1 focus:ring-primary/20 sm:w-72"
        />
      </div>

      {isLoading && (
        <LoadingState />
      )}
      {error && <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">{t('admin.loadFailed')}</p>}
      {data && data.items.length === 0 && (
        <EmptyState
          icon={<Users className="size-6 text-muted-foreground" />}
          title={t('admin.empty')}
          description={t('admin.usersEmptyDesc', { defaultValue: '用户注册或创建后，将在此处显示账户列表。' })}
        />
      )}

      {data && data.items.length > 0 && (
        <>
          <div className="space-y-3 md:hidden">
            {data.items.map((u) => (
              <div
                key={u.id}
                className="rounded-xl border border-border/50 bg-card/80 p-4 shadow-sm"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 space-y-1">
                    <p className="truncate text-sm font-semibold text-foreground">{u.name || '—'}</p>
                    <p className="truncate text-xs text-muted-foreground">{u.email}</p>
                  </div>
                  <div className="flex shrink-0 items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => openEdit(u)}
                      title={t('admin.edit')}
                    >
                      <Pencil className="size-4" />
                    </Button>
                    {u.role !== 'admin' && (
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => toggleStatus(u)}
                        disabled={saving === u.id}
                        title={u.status === 1 ? t('admin.freeze') : t('admin.activate')}
                      >
                        {u.status === 1 ? <Ban className="size-4" /> : <Unlock className="size-4" />}
                      </Button>
                    )}
                    {u.role !== 'admin' && (
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={() => setDeleteTarget(u.id)}
                        disabled={deleting === u.id}
                        className="text-destructive/70 hover:text-destructive"
                        title={t('admin.delete')}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    )}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">ID</p>
                    <p className="mt-1 font-medium text-foreground">{u.id}</p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.colRole')}</p>
                    <p className="mt-1">
                      <span className={['rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider', u.role === 'admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'].join(' ')}>
                        {u.role}
                      </span>
                    </p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.colStatus')}</p>
                    <p className="mt-1">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', u.status === 1 ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'].join(' ')}>
                        {u.status === 1 ? t('admin.active') : t('admin.frozen')}
                      </span>
                    </p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.colGroup', { defaultValue: '分组' })}</p>
                    <p className="mt-1 text-foreground">
                      {u.group_id ? (groups.find((g) => g.id === u.group_id)?.name || u.group_id) : '—'}
                    </p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.colImages')}</p>
                    <p className="mt-1 font-medium text-foreground">{u.image_num}</p>
                  </div>
                  <div className="rounded-lg bg-muted/20 px-2.5 py-2">
                    <p className="text-muted-foreground">{t('admin.usedCapacity', { defaultValue: '已用容量' })}</p>
                    <p className="mt-1 text-foreground">
                      {formatFileSize(u.used_capacity || 0)} / {u.capacity_bytes <= 0
                        ? t('settings.unlimitedCapacity')
                        : formatFileSize(u.capacity_bytes)}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="hidden overflow-x-auto rounded-xl border border-border/50 bg-card/80 shadow-sm md:block">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border/50 bg-muted/35 text-left text-xs text-muted-foreground">
                  <th className="px-4 py-3 font-medium">ID</th>
                  <th className="px-3 py-3 font-medium">{t('admin.colEmail')}</th>
                  <th className="px-3 py-3 font-medium">{t('admin.colName')}</th>
                  <th className="px-3 py-3 font-medium">
                    <div className="flex items-center gap-1">
                      {t('admin.colRole')}
                    </div>
                  </th>
                  <th className="px-3 py-3 font-medium">{t('admin.colGroup', { defaultValue: '分组' })}</th>
                  <th className="px-3 py-3 font-medium">{t('admin.colStatus')}</th>
                  <th className="px-3 py-3 font-medium">{t('admin.colImages')}</th>
                  <th className="px-3 py-3 font-medium">{t('admin.usedCapacity', { defaultValue: '已用容量' })}</th>
                  <th className="px-4 py-3 font-medium text-right">{t('admin.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {data.items.map((u) => (
                  <tr key={u.id} className="group hover:bg-muted/50 transition-colors">
                    <td className="px-4 py-3 text-muted-foreground">{u.id}</td>
                    <td className="px-3 py-3 text-foreground">{u.email}</td>
                    <td className="px-3 py-3 text-foreground">{u.name}</td>
                    <td className="px-3 py-3">
                      <span className={['rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider', u.role === 'admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'].join(' ')}>
                        {u.role}
                      </span>
                    </td>
                    <td className="px-3 py-3">
                      {u.group_id ? (
                        <span className="rounded px-1.5 py-0.5 text-xs font-medium bg-muted/50 text-foreground border border-border/50">
                          {groups.find(g => g.id === u.group_id)?.name || u.group_id}
                        </span>
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </td>
                    <td className="px-3 py-3">
                      <span className={['rounded px-1.5 py-0.5 text-xs font-medium', u.status === 1 ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'].join(' ')}>
                        {u.status === 1 ? t('admin.active') : t('admin.frozen')}
                      </span>
                    </td>
                    <td className="px-3 py-3 text-muted-foreground">{u.image_num}</td>
                    <td className="px-3 py-3 text-muted-foreground">
                      {formatFileSize(u.used_capacity || 0)} / {u.capacity_bytes <= 0
                        ? t('settings.unlimitedCapacity')
                        : formatFileSize(u.capacity_bytes)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => openEdit(u)}
                          title={t('admin.edit')}
                        >
                          <Pencil className="size-4" />
                        </Button>
                        {u.role !== 'admin' && (
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => toggleStatus(u)}
                            disabled={saving === u.id}
                            title={u.status === 1 ? t('admin.freeze') : t('admin.activate')}
                          >
                            {u.status === 1 ? <Ban className="size-4" /> : <Unlock className="size-4" />}
                          </Button>
                        )}
                        {u.role !== 'admin' && (
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => setDeleteTarget(u.id)}
                            disabled={deleting === u.id}
                            className="text-destructive/70 hover:text-destructive"
                            title={t('admin.delete')}
                          >
                            <Trash2 className="size-4" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {data.total > pageSize && (
            <div className="flex items-center justify-between pt-4">
              <span className="text-xs text-muted-foreground">{t('admin.pagination', { total: data.total })}</span>
              <div className="flex gap-2">
                <Button variant="outline" size="icon" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} title={t('admin.prev')}>
                  <ChevronLeft className="size-4" />
                </Button>
                <span className="inline-flex h-8 min-w-[56px] items-center justify-center rounded-lg border border-input bg-background px-2 text-xs text-muted-foreground">
                  {page} / {totalPages}
                </span>
                <Button variant="outline" size="icon" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} title={t('admin.next')}>
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('admin.confirmDeleteUser')}
        description={t('admin.deleteUserDescription')}
        confirmLabel={t('admin.delete')}
        onConfirm={onDelete}
        loading={!!deleting}
      />

      {/* Edit User Dialog */}
      <Dialog open={!!editing} onOpenChange={(open) => { if (!open) setEditing(null) }}>
        <DialogContent className="sm:max-w-[500px] border-border bg-card">
          <DialogHeader>
            <DialogTitle>{t('admin.editUser', { defaultValue: '编辑用户' })}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-5 pt-4 sm:grid-cols-2">
            
            <div className="space-y-3">
              <label className="text-sm font-medium text-foreground">{t('admin.colEmail')}</label>
              <input value={editing?.email || ''} disabled className="h-10 w-full rounded-lg border border-input bg-muted/50 px-3 text-sm text-muted-foreground cursor-not-allowed" />
            </div>

            <div className="space-y-3">
              <label className="text-sm font-medium text-foreground">{t('admin.colName')}</label>
              <input 
                value={editName} 
                onChange={(e) => setEditName(e.target.value)} 
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20" 
              />
            </div>

            <div className="space-y-3">
              <label className="text-sm font-medium text-foreground">{t('admin.colGroup', { defaultValue: '所属分组' })}</label>
              <Select
                value={editGroup}
                onValueChange={(val) => val !== null && setEditGroup(val as string)}
                items={{
                  none: t('admin.noGroup', { defaultValue: '不分配分组' }),
                  ...Object.fromEntries(groups.map((g) => [g.id.toString(), g.name])),
                }}
              >
                <SelectTrigger className="h-10 w-full">
                  <SelectValue placeholder={t('admin.noGroup', { defaultValue: '不分配分组' })} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">{t('admin.noGroup', { defaultValue: '不分配分组' })}</SelectItem>
                  {groups.map(g => (
                    <SelectItem key={g.id} value={g.id.toString()}>{g.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-3">
              <label className="text-sm font-medium text-foreground">{t('admin.capacity', { defaultValue: '存储容量' })}</label>
              <div className="flex items-center gap-2">
                <input 
                  type="number" 
                  min={0} 
                  value={editCapacity} 
                  onChange={(e) => setEditCapacity(Number(e.target.value))} 
                  className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20" 
                />
                <span className="text-sm text-muted-foreground">MB</span>
              </div>
            </div>

            <div className="space-y-3 sm:col-span-2">
              <label className="text-sm font-medium text-foreground">{t('admin.passwordReset', { defaultValue: '重置密码 (留空则不修改)' })}</label>
              <input 
                type="password"
                value={editPassword} 
                onChange={(e) => setEditPassword(e.target.value)} 
                placeholder="******"
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none transition-colors duration-150 focus:border-primary focus:ring-2 focus:ring-primary/20" 
              />
            </div>
          </div>
          <div className="mt-3 flex justify-end gap-3 border-t border-border pt-5">
            <Button variant="outline" size="lg" onClick={() => setEditing(null)}>
              {t('admin.cancel')}
            </Button>
            <Button size="lg" onClick={handleEditSave} disabled={editSaving}>
              {editSaving ? '…' : t('admin.confirmSave')}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </section>
  )
}
