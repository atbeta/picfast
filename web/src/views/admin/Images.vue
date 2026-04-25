<template>
  <div>
    <n-h3 class="mb-4">Image Management</n-h3>
    <n-data-table :columns="columns" :data="images" :loading="loading" :pagination="pagination" remote
      @update:page="fetchImages" />
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NH3, NDataTable, NTag, NButton, NImage, NSpace, useMessage, useDialog, type DataTableColumns } from 'naive-ui'
import { adminGetImages, adminDeleteImage } from '../../api/admin'

const message = useMessage()
const dialog = useDialog()
const images = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const columns: DataTableColumns = [
  {
    title: 'Preview', key: 'preview', width: 80,
    render: (row: any) => h(NImage, { src: row.thumbnail_url || row.url, width: 50, height: 50, objectFit: 'cover' }),
  },
  { title: 'Key', key: 'key', width: 120 },
  { title: 'User', key: 'user_name', width: 100, render: (row: any) => row.user_name || `#${row.user_id}` },
  { title: 'MIME', key: 'mime_type', width: 100 },
  {
    title: 'Size', key: 'size_bytes', width: 90,
    render: (row: any) => formatSize(row.size_bytes),
  },
  {
    title: 'Public', key: 'permission', width: 70,
    render: (row: any) => h(NTag, { type: row.permission === 1 ? 'success' : 'warning', size: 'small' }, () => row.permission === 1 ? 'Yes' : 'No'),
  },
  {
    title: 'Actions', key: 'actions', width: 80,
    render: (row: any) =>
      h(NButton, { size: 'small', type: 'error', onClick: () => confirmDelete(row) }, () => 'Delete'),
  },
]

onMounted(() => fetchImages())

async function fetchImages(page = 1) {
  loading.value = true
  try {
    const res = await adminGetImages({ page: String(page), page_size: '20' })
    images.value = res.data.data
    pagination.value.itemCount = res.data.pagination?.total || 0
    pagination.value.page = page
  } catch {
    message.error('Failed to load images')
  } finally {
    loading.value = false
  }
}

function confirmDelete(img: any) {
  dialog.warning({
    title: 'Delete Image',
    content: `Delete image "${img.key}"?`,
    positiveText: 'Delete',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        await adminDeleteImage(img.id)
        message.success('Image deleted')
        fetchImages(pagination.value.page)
      } catch {
        message.error('Failed to delete image')
      }
    },
  })
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}
</script>
