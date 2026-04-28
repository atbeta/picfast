import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { createAlbum, deleteAlbum, listAlbums, listImages, updateAlbum } from '../../lib/console-api'
import type { ImageItem } from '../../lib/console-api'
import { ConfirmDialog } from '@/components/confirm-dialog'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

function toRelative(url: string): string {
  try { return new URL(url).pathname }
  catch { return url }
}

export function AlbumsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
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

  // Album image viewer
  const [viewingAlbum, setViewingAlbum] = useState<{ id: number; name: string } | null>(null)
  const [albumImages, setAlbumImages] = useState<ImageItem[]>([])
  const [albumImagesLoading, setAlbumImagesLoading] = useState(false)

  const viewAlbumImages = async (album: { id: number; name: string }) => {
    setViewingAlbum(album)
    setAlbumImagesLoading(true)
    try {
      const res = await listImages(1, 50)
      const all = res.items || []
      setAlbumImages(all.filter((img) => img.album_id === album.id))
    } catch {
      setAlbumImages([])
    } finally {
      setAlbumImagesLoading(false)
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t('page.albums.title')}</h1>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-zinc-900 px-3 py-1.5 text-sm text-white hover:bg-zinc-700 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-300"
        >
          {t('albums.create')}
        </button>
      </div>

      {showCreate && (
        <div className="space-y-3 rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-800">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder={t('albums.namePlaceholder')}
            className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900"
          />
          <input
            value={newIntro}
            onChange={(e) => setNewIntro(e.target.value)}
            placeholder={t('albums.introPlaceholder')}
            className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-900"
          />
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleCreate}
              disabled={creating || !newName.trim()}
              className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
            >
              {creating ? t('albums.creating') : t('albums.confirmCreate')}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName(''); setNewIntro('') }}
              className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600"
            >
              {t('albums.cancel')}
            </button>
          </div>
        </div>
      )}

      {isLoading && (
        <div className="flex justify-center py-12">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
        </div>
      )}
      {error && (
        <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {t('albums.loadFailed')}
        </p>
      )}
      {data && data.items.length === 0 && (
        <p className="py-12 text-center text-sm text-zinc-400">{t('albums.empty')}</p>
      )}

      {/* Card grid */}
      {data && data.items.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data.items.map((album) =>
            editingId === album.id ? (
              <div key={album.id} className="space-y-3 rounded-lg border border-zinc-200 p-4 dark:border-zinc-700">
                <input
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-800"
                />
                <input
                  value={editIntro}
                  onChange={(e) => setEditIntro(e.target.value)}
                  className="w-full rounded-md border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 dark:border-zinc-600 dark:bg-zinc-800"
                />
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleUpdate}
                    disabled={saving}
                    className="rounded-md bg-zinc-900 px-3 py-1.5 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
                  >
                    {saving ? t('albums.saving') : t('albums.confirmSave')}
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditingId(null)}
                    className="rounded-md border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600"
                  >
                    {t('albums.cancel')}
                  </button>
                </div>
              </div>
            ) : (
              <div key={album.id} className="rounded-lg border border-zinc-200 p-4 transition-shadow hover:shadow-md dark:border-zinc-700">
                <div className="mb-2 flex items-center justify-between">
                  <button
                    type="button"
                    onClick={() => viewAlbumImages(album)}
                    className="text-left text-sm font-medium text-zinc-800 hover:text-blue-600 dark:text-zinc-200 dark:hover:text-blue-400"
                  >
                    {album.name}
                  </button>
                  <div className="flex gap-1">
                    <button type="button" onClick={() => startEdit(album)} className="rounded px-2 py-1 text-xs hover:bg-zinc-100 dark:hover:bg-zinc-800">{t('albums.edit')}</button>
                    <button type="button" onClick={() => setDeleteTarget(album.id)} disabled={deleting === album.id} className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50 disabled:opacity-50 dark:hover:bg-red-900/20">{t('albums.delete')}</button>
                  </div>
                </div>
                <p className="text-xs text-zinc-400">{album.intro || t('albums.noDesc', { defaultValue: '暂无描述' })}</p>
                <p className="mt-2 border-t border-zinc-100 pt-2 text-xs text-zinc-400 dark:border-zinc-800">
                  {t('albums.imageCount', { count: album.image_num })} · {new Date(album.created_at).toLocaleDateString()}
                </p>
              </div>
            ),
          )}
        </div>
      )}

      {data && data.total > 20 && (
        <div className="flex items-center justify-between pt-2">
          <span className="text-xs text-zinc-500">{t('albums.pagination', { total: data.total })}</span>
          <div className="flex gap-2">
            <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('albums.prev')}</button>
            <button type="button" disabled={page * 20 >= data.total} onClick={() => setPage((p) => p + 1)} className="rounded border border-zinc-300 px-3 py-1 text-xs disabled:opacity-40 dark:border-zinc-700">{t('albums.next')}</button>
          </div>
        </div>
      )}

      {/* Album image viewer dialog */}
      <Dialog open={!!viewingAlbum} onOpenChange={(open) => { if (!open) setViewingAlbum(null) }}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{viewingAlbum?.name} - {t('albums.imageList', { defaultValue: '图片列表' })}</DialogTitle>
          </DialogHeader>

          {albumImagesLoading && <div className="flex justify-center py-8"><div className="h-6 w-6 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" /></div>}

          {!albumImagesLoading && albumImages.length === 0 && (
            <p className="py-8 text-center text-sm text-zinc-400">{t('albums.noImages', { defaultValue: '相册内暂无图片' })}</p>
          )}

          {!albumImagesLoading && albumImages.length > 0 && (
            <div className="grid grid-cols-4 gap-2">
              {albumImages.map((img) => (
                <div key={img.key} className="aspect-square overflow-hidden rounded">
                  {img.thumbnail_url ? (
                    <img src={toRelative(img.thumbnail_url)} alt="" className="h-full w-full object-cover" loading="lazy" />
                  ) : img.url ? (
                    <img src={toRelative(img.url)} alt="" className="h-full w-full object-cover" loading="lazy" />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center bg-zinc-100 text-xs text-zinc-400 dark:bg-zinc-900">{img.extension}</div>
                  )}
                </div>
              ))}
            </div>
          )}
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={t('albums.confirmDelete')}
        destructive
        confirmLabel={t('albums.delete')}
        onConfirm={handleDelete}
        loading={!!deleting}
      />
    </section>
  )
}
