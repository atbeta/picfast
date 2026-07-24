import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Tag as TagIcon } from 'lucide-react'
import { toast } from 'sonner'

import { listTags, createTag } from '@/lib/console-api'
import { Button } from '@/components/ui/button'
import { EmptyState, LoadingState } from '@/components/page-states'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'

export function TagsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [createOpen, setCreateOpen] = useState(false)
  const [newTagName, setNewTagName] = useState('')

  const { data: tags = [], isLoading: loading, isError } = useQuery({
    queryKey: ['tags'],
    queryFn: listTags,
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => createTag(name),
    onSuccess: () => {
      toast.success(t('tags.createSuccess'))
      setCreateOpen(false)
      setNewTagName('')
      qc.invalidateQueries({ queryKey: ['tags'] })
    },
    onError: () => {
      toast.error(t('tags.createFailed'))
    },
  })

  const handleCreate = () => {
    if (!newTagName.trim()) return
    createMutation.mutate(newTagName.trim())
  }

  if (isError) {
    toast.error(t('tags.loadFailed'))
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">
          {t('tags.title')}
        </h1>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 size-4" />
          {t('tags.create')}
        </Button>
      </div>

      {loading ? (
        <LoadingState />
      ) : tags.length === 0 ? (
        <EmptyState
          icon={<TagIcon className="size-6 text-muted-foreground" />}
          title={t('tags.empty')}
          description={t('tags.emptyDesc')}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          {tags.map((tag) => (
            <div key={tag.id} className="flex flex-col gap-2 rounded-xl border bg-card p-4 transition-colors hover:border-primary/30">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <TagIcon className="size-4 text-primary" />
                  <span className="font-medium text-foreground">{tag.name}</span>
                </div>
                <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                  {tag.type}
                </span>
              </div>
              <div className="mt-2 text-xs text-muted-foreground">
                {t('tags.createdAt')}: {new Date(tag.created_at).toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('tags.create')}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label htmlFor="name" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">{t('tags.name')}</label>
              <input
                id="name"
                value={newTagName}
                onChange={(e) => setNewTagName(e.target.value)}
                placeholder={t('tags.namePlaceholder')}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={createMutation.isPending}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleCreate} disabled={!newTagName.trim() || createMutation.isPending}>
              {t('common.confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
