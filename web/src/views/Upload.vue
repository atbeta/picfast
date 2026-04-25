<template>
  <div>
    <n-h2>Upload Images</n-h2>
    <n-upload
      multiple
      directory-dnd
      :custom-request="handleUpload"
      :show-file-list="false"
      accept="image/*"
    >
      <n-upload-dragger>
        <div class="py-8">
          <n-icon size="48" :depth="3"><cloud-upload-outline /></n-icon>
          <n-text class="block mt-2">Click or drag files here to upload</n-text>
          <n-p depth="3" class="mt-1">Supports JPG, PNG, GIF, WebP, BMP, SVG</n-p>
        </div>
      </n-upload-dragger>
    </n-upload>

    <div v-if="results.length" class="mt-6">
      <n-h3>Upload Results</n-h3>
      <n-grid :cols="1" :x-gap="12" :y-gap="12">
        <n-gi v-for="(item, i) in results" :key="i">
          <n-card size="small">
            <div class="flex items-center gap-4">
              <n-image :src="item.thumbnail_url || item.url" width="80" height="80" object-fit="cover" />
              <div class="flex-1 min-w-0">
                <n-text strong>{{ item.key }}</n-text>
                <n-input-group class="mt-1">
                  <n-input :value="item.links?.url" size="small" readonly />
                  <n-button size="small" @click="copyText(item.links?.url)">Copy</n-button>
                </n-input-group>
                <div class="flex gap-2 mt-1">
                  <n-button size="tiny" @click="copyText(item.links?.markdown)">Markdown</n-button>
                  <n-button size="tiny" @click="copyText(item.links?.bbcode)">BBCode</n-button>
                  <n-button size="tiny" @click="copyText(item.links?.html)">HTML</n-button>
                </div>
              </div>
              <n-tag size="small" :type="item.status === 'ok' ? 'success' : 'error'">
                {{ item.status === 'ok' ? 'Success' : 'Failed' }}
              </n-tag>
            </div>
          </n-card>
        </n-gi>
      </n-grid>
      <n-button class="mt-4" @click="results = []">Clear Results</n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import {
  NH2, NH3, NUpload, NUploadDragger, NIcon, NText, NP, NGrid, NGi, NCard,
  NImage, NInputGroup, NInput, NButton, NTag, useMessage,
} from 'naive-ui'
import { CloudUploadOutline } from '@vicons/ionicons5'
import { uploadImage } from '../api/image'

const message = useMessage()
const results = ref<any[]>([])

function copyText(text?: string) {
  if (!text) return
  navigator.clipboard.writeText(text).then(() => message.success('Copied')).catch(() => {})
}

async function handleUpload({ file }: any) {
  try {
    const res = await uploadImage(file.file)
    const data = res.data.data
    results.value.unshift({ ...data, status: 'ok' })
    message.success(`Uploaded: ${data.key}`)
  } catch (err: any) {
    results.value.unshift({ key: file.name, status: 'error', message: err.response?.data?.message || 'Upload failed' })
    message.error(`Failed: ${file.name}`)
  }
}
</script>
