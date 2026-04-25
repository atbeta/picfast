<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <n-h2 class="m-0">My Images</n-h2>
    </div>

    <n-spin :show="loading">
      <n-grid :cols="4" :x-gap="12" :y-gap="12" responsive="screen">
        <n-gi v-for="img in images" :key="img.key">
          <n-card size="small" hoverable @click="showDetail(img)">
            <template #cover>
              <n-image
                :src="img.thumbnail_url || img.url"
                width="100%"
                height="160"
                object-fit="cover"
                preview-disabled
              />
            </template>
            <div class="flex items-center justify-between">
              <n-text depth="3" class="text-xs truncate">{{ img.key }}</n-text>
              <n-tag :type="img.permission === 1 ? 'success' : 'warning'" size="tiny">
                {{ img.permission === 1 ? 'Public' : 'Private' }}
              </n-tag>
            </div>
            <n-text depth="3" class="text-xs">{{ formatSize(img.size_bytes) }} · {{ img.mime_type }}</n-text>
          </n-card>
        </n-gi>
      </n-grid>
      <n-empty v-if="!loading && images.length === 0" description="No images yet" class="mt-12" />
    </n-spin>

    <div class="flex justify-center mt-6" v-if="total > pageSize">
      <n-pagination v-model:page="page" :page-count="Math.ceil(total / pageSize)" @update:page="fetchImages" />
    </div>

    <n-modal v-model:show="detailVisible" preset="card" title="Image Detail" style="width: 600px">
      <template v-if="detail">
        <div class="flex justify-center mb-4">
          <n-image :src="detail.url" :alt="detail.key" style="max-height: 300px" />
        </div>
        <n-descriptions bordered :column="1" label-placement="left" size="small">
          <n-descriptions-item label="Key">{{ detail.key }}</n-descriptions-item>
          <n-descriptions-item label="Size">{{ formatSize(detail.size_bytes) }}</n-descriptions-item>
          <n-descriptions-item label="MIME">{{ detail.mime_type }}</n-descriptions-item>
          <n-descriptions-item label="Dimensions">{{ detail.width }}x{{ detail.height }}</n-descriptions-item>
          <n-descriptions-item label="Permission">
            <n-switch :value="detail.permission === 1" @update:value="togglePermission" />
          </n-descriptions-item>
        </n-descriptions>
        <div class="mt-4 space-y-2" v-if="detail.links">
          <n-input-group v-for="(val, fmt) in detail.links" :key="fmt">
            <n-input :value="val" size="small" readonly />
            <n-button size="small" @click="copyText(val)">{{ fmt }}</n-button>
          </n-input-group>
        </div>
        <div class="flex justify-end mt-4">
          <n-button type="error" @click="deleteDetail">Delete</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NH2, NGrid, NGi, NCard, NImage, NText, NTag, NEmpty, NPagination, NSpin,
  NModal, NDescriptions, NDescriptionsItem, NSwitch, NInputGroup, NInput, NButton, useMessage,
} from 'naive-ui'
import { getImages, getImage, deleteImage, updateImage } from '../api/image'

const message = useMessage()
const images = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const detailVisible = ref(false)
const detail = ref<any>(null)

onMounted(() => fetchImages())

async function fetchImages() {
  loading.value = true
  try {
    const res = await getImages(page.value, pageSize)
    images.value = res.data.data
    total.value = res.data.pagination?.total || 0
  } catch {
    message.error('Failed to load images')
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
    message.error('Failed to load image detail')
  }
}

async function togglePermission(public_: boolean) {
  if (!detail.value) return
  try {
    await updateImage(detail.value.key, { permission: public_ ? 1 : 0 })
    detail.value.permission = public_ ? 1 : 0
    fetchImages()
  } catch {
    message.error('Failed to update permission')
  }
}

async function deleteDetail() {
  if (!detail.value) return
  try {
    await deleteImage(detail.value.key)
    message.success('Image deleted')
    detailVisible.value = false
    fetchImages()
  } catch {
    message.error('Failed to delete image')
  }
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => message.success('Copied')).catch(() => {})
}
</script>
