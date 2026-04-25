<template>
  <div>
    <n-h3 style="margin-bottom: 16px;">图片管理</n-h3>
    <n-data-table :columns="columns" :data="images" :loading="loading" :pagination="pagination" remote
      @update:page="fetchImages" />
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue'
import { NH3, NDataTable, NTag, NButton, NImage, useMessage, useDialog, type DataTableColumns } from 'naive-ui'
import { adminGetImages, adminDeleteImage } from '../../api/admin'

const message = useMessage()
const dialog = useDialog()
const images = ref<any[]>([])
const loading = ref(false)
const pagination = ref({ page: 1, pageSize: 20, itemCount: 0 })

const columns: DataTableColumns = [
  {
    title: '预览', key: 'preview', width: 70,
    render: (row: any) => h(NImage, { src: row.thumbnail_url || row.url, width: 48, height: 48, objectFit: 'cover', previewDisabled: true }),
  },
  { title: 'Key', key: 'key', width: 100 },
  { title: '用户', key: 'user_name', width: 80, render: (row: any) => row.user_name || `#${row.user_id}` },
  { title: '文件名', key: 'origin_name', ellipsis: { tooltip: true } },
  { title: '大小', key: 'size_bytes', width: 80, render: (row: any) => formatSize(row.size_bytes) },
  { title: '权限', key: 'permission', width: 60, render: (row: any) => h(NTag, { type: row.permission === 1 ? 'success' : 'warning', size: 'small' }, () => row.permission === 1 ? '公开' : '私有') },
  {
    title: '操作', key: 'actions', width: 70,
    render: (row: any) => h(NButton, { size: 'small', type: 'error', onClick: () => confirmDelete(row) }, () => '删除'),
  },
]

onMounted(() => fetchImages())

async function fetchImages(page = 1) {
  loading.value = true
  try {
    const res = await adminGetImages({ page: String(page), page_size: '20' })
    const d = res.data.data
    images.value = d.items || d
    pagination.value.itemCount = d.total || res.data.pagination?.total || 0
    pagination.value.page = page
  } catch {
    message.error('加载图片列表失败')
  } finally {
    loading.value = false
  }
}

function confirmDelete(img: any) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除图片 "${img.key}" 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminDeleteImage(img.id)
        message.success('图片已删除')
        fetchImages(pagination.value.page)
      } catch {
        message.error('删除失败')
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
