<template>
  <div>
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <n-h2 style="margin: 0;">我的图片</n-h2>
      <n-text depth="3">共 {{ total }} 张</n-text>
    </div>

    <n-spin :show="loading">
      <n-grid :cols="4" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
        <n-gi v-for="img in images" :key="img.key" span="4 sm:2 md:1">
          <n-card size="small" hoverable @click="showDetail(img)" style="cursor: pointer;">
            <template #cover>
              <div style="height: 160px; overflow: hidden; background: #f5f5f5; display: flex; align-items: center; justify-content: center;">
                <n-image
                  :src="imgSrc(img)"
                  width="100%"
                  height="160"
                  object-fit="cover"
                  preview-disabled
                  fallback-src=""
                >
                  <template #error>
                    <n-text depth="3" style="font-size: 12px;">无预览</n-text>
                  </template>
                </n-image>
              </div>
            </template>
            <div style="display: flex; align-items: center; justify-content: space-between;">
              <n-text depth="3" style="font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 120px;">{{ img.origin_name }}</n-text>
              <n-tag :type="img.permission === 1 ? 'success' : 'warning'" size="tiny">
                {{ img.permission === 1 ? '公开' : '私有' }}
              </n-tag>
            </div>
            <n-text depth="3" style="font-size: 12px;">{{ formatSize(img.size_bytes) }} · {{ img.width || 0 }}x{{ img.height || 0 }}</n-text>
          </n-card>
        </n-gi>
      </n-grid>
      <n-empty v-if="!loading && images.length === 0" description="暂无图片，去上传吧" style="margin-top: 48px;" />
    </n-spin>

    <div style="display: flex; justify-content: center; margin-top: 24px;" v-if="total > pageSize">
      <n-pagination v-model:page="page" :page-count="Math.ceil(total / pageSize)" @update:page="fetchImages" />
    </div>

    <n-modal v-model:show="detailVisible" preset="card" title="图片详情" style="width: 600px;">
      <template v-if="detail">
        <div style="display: flex; justify-content: center; margin-bottom: 16px; background: #f9f9f9; border-radius: 8px; padding: 8px;">
          <n-image :src="detail.url" :alt="detail.key" style="max-height: 300px;" />
        </div>
        <n-descriptions bordered :column="2" label-placement="left" size="small">
          <n-descriptions-item label="Key">{{ detail.key }}</n-descriptions-item>
          <n-descriptions-item label="文件名">{{ detail.origin_name }}</n-descriptions-item>
          <n-descriptions-item label="大小">{{ formatSize(detail.size_bytes) }}</n-descriptions-item>
          <n-descriptions-item label="类型">{{ detail.mimetype }}</n-descriptions-item>
          <n-descriptions-item label="尺寸">{{ detail.width }}x{{ detail.height }}</n-descriptions-item>
          <n-descriptions-item label="权限">
            <n-switch :value="detail.permission === 1" @update:value="togglePermission">
              <template #checked>公开</template>
              <template #unchecked>私有</template>
            </n-switch>
          </n-descriptions-item>
        </n-descriptions>
        <div style="margin-top: 16px;" v-if="detail.links">
          <n-input-group v-for="(val, fmt) in detail.links" :key="fmt" style="margin-bottom: 8px;">
            <n-input :value="val" size="small" readonly />
            <n-button size="small" @click="copyText(val)">{{ fmt }}</n-button>
          </n-input-group>
        </div>
        <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
          <n-button type="error" @click="deleteDetail">删除图片</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NH2, NGrid, NGi, NCard, NImage, NText, NTag, NEmpty, NPagination, NSpin,
  NModal, NDescriptions, NDescriptionsItem, NSwitch, NInputGroup, NInput, NButton, useMessage, useDialog,
} from 'naive-ui'
import { getImages, getImage, deleteImage, updateImage } from '../api/image'

const message = useMessage()
const dialog = useDialog()
const images = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const detailVisible = ref(false)
const detail = ref<any>(null)

onMounted(() => fetchImages())

function imgSrc(img: any) {
  return img.thumbnail_url || img.url || ''
}

async function fetchImages() {
  loading.value = true
  try {
    const res = await getImages(page.value, pageSize)
    const d = res.data.data
    images.value = d.items || d
    total.value = d.total || 0
  } catch {
    message.error('加载图片失败')
  } finally {
    loading.value = false
  }
}

async function showDetail(img: any) {
  try {
    const res = await getImage(img.key)
    detail.value = res.data.data
    detailVisible.value = true
  } catch {
    message.error('加载图片详情失败')
  }
}

async function togglePermission(public_: boolean) {
  if (!detail.value) return
  try {
    await updateImage(detail.value.key, { permission: public_ ? 1 : 0 })
    detail.value.permission = public_ ? 1 : 0
    message.success(public_ ? '已设为公开' : '已设为私有')
    fetchImages()
  } catch {
    message.error('修改权限失败')
  }
}

async function deleteDetail() {
  if (!detail.value) return
  dialog.warning({
    title: '确认删除',
    content: `确定要删除图片 "${detail.value.origin_name}" 吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteImage(detail.value.key)
        message.success('图片已删除')
        detailVisible.value = false
        fetchImages()
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

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => message.success('已复制')).catch(() => {})
}
</script>
