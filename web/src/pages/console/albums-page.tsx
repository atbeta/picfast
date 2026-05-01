import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Plus, Pencil, Trash2, Image as ImageIcon, ChevronLeft, ChevronRight } from 'lucide-react'

import { createAlbum, deleteAlbum, listAlbums, updateAlbum } from '../../lib/console-api'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { EmptyState, LoadingState } from '@/components/page-states'

export function AlbumsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const navigate = useNavigate()
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['albums', page],
    queryFn: () => listAlbums(page, 20),
  })

  // Create
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newIntro, setNewIntro] = useState('')
  const [creating, setCreating] = useState(false)

  const handleCreate = useCallback(async () => {
    if (!newName.trim()) return
    setCreating(true)
    try {
      await createAlbum(newName.trim(), newIntro.trim() || undefined)
      setShowCreate(false)
      setNewName('')
      setNewIntro('')
      await qc.invalidateQueries({ queryKey: ['albums'] })
    } catch (err: unknown) {
      toast.error(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('albums.createFailed'),
      )
    } finally {
      setCreating(false)
    }
  }, [newName, newIntro, qc, t])

  // Edit
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editIntro, setEditIntro] = useState('')
  const [saving, setSaving] = useState(false)

  const startEdit = (album: { id: number; name: string; intro: string }) => {
    setEditingId(album.id)
    setEditName(album.name)
    setEditIntro(album.intro)
  }

  const handleUpdate = useCallback(async () => {
    if (!editingId || !editName.trim()) return
    setSaving(true)
    try {
      await updateAlbum(editingId, { name: editName.trim(), intro: editIntro.trim() || undefined })
      setEditingId(null)
      await qc.invalidateQueries({ queryKey: ['albums'] })
    } catch (err: unknown) {
      toast.error(
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
        t('albums.updateFailed'),
      )
    } finally {
      setSaving(false)
    }
  }, [editingId, editName, editIntro, qc, t])

  // Delete
  const [deleting, setDeleting] = useState<number | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<number | null>(null)

  const handleDelete = useCallback(
    async () => {
      if (deleteTarget === null) return
      setDeleting(deleteTarget)
      try {
        await deleteAlbum(deleteTarget)
        setDeleteTarget(null)
        await qc.invalidateQueries({ queryKey: ['albums'] })
      } catch (err: unknown) {
        toast.error(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          t('albums.deleteFailed'),
        )
      } finally {
        setDeleting(null)
      }
    },
    [deleteTarget, qc, t],
  )

  const viewAlbumImages = (album: { id: number; name: string }) => {
    navigate(`/console/images?album_id=${album.id}`)
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">{t('page.albums.title')}</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-all hover:bg-primary/90 hover:scale-105 active:scale-95 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring cursor-pointer"
        >
          <Plus className="-ms-1 me-2 size-4" />
          {t('albums.create')}
        </button>
      </div>

      {showCreate && (
        <div className="rounded-xl border border-border/50 bg-card p-6 shadow-sm">
          <div className="mb-4">
            <h3 className="text-lg font-medium leading-none tracking-tight">{t('albums.create')}</h3>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder={t('albums.namePlaceholder')}
                className="flex h-10 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
            <div className="space-y-2">
              <input
                value={newIntro}
                onChange={(e) => setNewIntro(e.target.value)}
                placeholder={t('albums.introPlaceholder')}
                className="flex h-10 w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <button
              type="button"
              onClick={handleCreate}
              disabled={creating || !newName.trim()}
              className="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer"
            >
              {creating ? t('albums.creating') : t('albums.confirmCreate')}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName(''); setNewIntro('') }}
              className="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-lg border border-input bg-background px-4 py-2 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer"
            >
              {t('albums.cancel')}
            </button>
          </div>
        </div>
      )}

      {isLoading && (
        <LoadingState />
      )}
      {error && (
        <p className="rounded-xl border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive">
          {t('albums.loadFailed')}
        </p>
      )}
      {data && data.items.length === 0 && (
        <EmptyState
          icon={<ImageIcon className="size-8 text-muted-foreground/60" />}
          title={t('albums.empty')}
          description={t('albums.emptyDesc', { defaultValue: '创建相册后，就可以更轻松地整理图片。' })}
          className="min-h-[300px]"
        />
      )}

      {/* Album card grid */}
      {data && data.items.length > 0 && (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {data.items.map((album) =>
            editingId === album.id ? (
              <div key={album.id} className="rounded-xl border border-primary/30 bg-card p-5 shadow-md transition-all">
                <div className="space-y-3">
                  <input
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  />
                  <input
                    value={editIntro}
                    onChange={(e) => setEditIntro(e.target.value)}
                    className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                  />
                  <div className="flex gap-2 pt-2">
                    <button
                      type="button"
                      onClick={handleUpdate}
                      disabled={saving}
                      className="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer"
                    >
                      {saving ? t('albums.saving') : t('albums.confirmSave')}
                    </button>
                    <button
                      type="button"
                      onClick={() => setEditingId(null)}
                      className="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-lg border border-input bg-background px-4 py-2 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer"
                    >
                      {t('albums.cancel')}
                    </button>
                  </div>
                </div>
              </div>
            ) : (
              <div key={album.id} className="group relative flex flex-col justify-between rounded-xl border border-border/50 bg-card p-4 shadow-sm transition-all hover:shadow-md hover:-translate-y-1 hover:border-primary/30">
                <div 
                  className="mb-4 aspect-video w-full overflow-hidden rounded-lg bg-muted cursor-pointer relative"
                  onClick={() => viewAlbumImages(album)}
                >
                  {album.cover_md5 ? (
                    <img 
                      src={`/t/${album.cover_md5}.png`} 
                      alt={album.name} 
                      className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" 
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center">
                      <ImageIcon className="size-8 text-muted-foreground/30" />
                    </div>
                  )}
                  <div className="absolute inset-0 bg-black/0 transition-colors duration-300 group-hover:bg-black/10" />
                </div>
                
                <div className="mb-4 flex items-start justify-between">
                  <div className="space-y-1.5 pe-4">
                    <button
                      type="button"
                      onClick={() => viewAlbumImages(album)}
                      className="text-left font-semibold text-foreground transition-colors hover:text-primary cursor-pointer"
                    >
                      {album.name}
                    </button>
                    <p className="line-clamp-2 text-sm text-muted-foreground">
                      {album.intro || t('albums.noDesc', { defaultValue: '暂无描述' })}
                    </p>
                  </div>
                  <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <button 
                      type="button" 
                      onClick={() => startEdit(album)} 
                      className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground cursor-pointer"
                      title={t('albums.edit')}
                    >
                      <Pencil className="size-4" />
                    </button>
                    <button 
                      type="button" 
                      onClick={() => setDeleteTarget(album.id)} 
                      disabled={deleting === album.id} 
                      className="rounded-lg p-1.5 text-destructive/70 transition-colors hover:bg-destructive/10 hover:text-destructive disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
                      title={t('albums.delete')}
                    >
                      <Trash2 className="size-4" />
                    </button>
                  </div>
                </div>
                <div className="mt-auto flex items-center justify-between border-t border-border/50 pt-4 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1.5">
                    <ImageIcon className="size-3.5" />
                    {t('albums.imageCount', { count: album.image_num })}
                  </span>
                  <span>{new Date(album.created_at).toLocaleDateString()}</span>
                </div>
              </div>
            ),
          )}
        </div>
      )}

      {data && data.total > 20 && (
        <div className="flex items-center justify-between border-t border-border/50 pt-4">
          <span className="text-sm text-muted-foreground">{t('albums.pagination', { total: data.total })}</span>
          <div className="flex gap-2">
            <button 
              type="button" 
              disabled={page <= 1} 
              onClick={() => setPage((p) => p - 1)} 
              title={t('albums.prev')}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
            >
              <ChevronLeft className="size-4" />
            </button>
            <button 
              type="button" 
              disabled={page * 20 >= data.total} 
              onClick={() => setPage((p) => p + 1)} 
              title={t('albums.next')}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-input bg-background shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
            >
              <ChevronRight className="size-4" />
            </button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('albums.confirmDelete')}
        description={t('albums.deleteDescription')}
        confirmLabel={t('albums.delete')}
        onConfirm={handleDelete}
        loading={!!deleting}
      />
    </section>
  )
}
