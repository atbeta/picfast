<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <n-h3 class="m-0">Storage Strategies</n-h3>
      <n-button type="primary" @click="openCreate">New Strategy</n-button>
    </div>
    <n-data-table :columns="columns" :data="strategies" :loading="loading" />

    <n-modal v-model:show="showModal" preset="dialog" :title="editing ? 'Edit Strategy' : 'Create Strategy'"
      positive-text="Save" negative-text="Cancel" @positive-click="saveStrategy" style="width: 550px">
      <n-form>
        <n-form-item label="Name">
          <n-input v-model:value="form.name" />
        </n-form-item>
        <n-form-item label="Type">
          <n-select v-model:value="form.type" :options="typeOptions" :disabled="!!editing" />
        </n-form-item>
        <template v-if="form.type === 'local'">
          <n-form-item label="Root Path">
            <n-input v-model:value="form.localRoot" placeholder="/data/images" />
          </n-form-item>
          <n-form-item label="URL Prefix">
            <n-input v-model:value="form.localUrl" placeholder="http://localhost:8080/i" />
          </n-form-item>
        </template>
        <template v-if="form.type === 's3'">
          <n-form-item label="Endpoint">
            <n-input v-model:value="form.s3Endpoint" placeholder="https://s3.amazonaws.com" />
          </n-form-item>
          <n-form-item label="Region">
            <n-input v-model:value="form.s3Region" placeholder="us-east-1" />
          </n-form-item>
          <n-form-item label="Bucket">
            <n-input v-model:value="form.s3Bucket" />
          </n-form-item>
          <n-form-item label="Access Key">
            <n-input v-model:value="form.s3AccessKey" />
          </n-form-item>
          <n-form-item label="Secret Key">
            <n-input v-model:value="form.s3SecretKey" type="password" show-password-on="click" />
          </n-form-item>
          <n-form-item label="URL Prefix">
            <n-input v-model:value="form.s3Url" placeholder="https://bucket.s3.amazonaws.com" />
          </n-form-item>
        </template>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NH3, NDataTable, NButton, NSpace, NModal, NForm, NFormItem, NInput, NSelect, useMessage, type DataTableColumns } from 'naive-ui'
import { adminGetStrategies, adminCreateStrategy, adminUpdateStrategy, adminDeleteStrategy } from '../../api/admin'

const message = useMessage()
const strategies = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref<any>(null)
const form = reactive({
  name: '', type: 'local' as string,
  localRoot: '/data/images', localUrl: '',
  s3Endpoint: '', s3Region: 'us-east-1', s3Bucket: '', s3AccessKey: '', s3SecretKey: '', s3Url: '',
})

const typeOptions = [
  { label: 'Local', value: 'local' },
  { label: 'S3 Compatible', value: 's3' },
]

const columns: DataTableColumns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: 'Name', key: 'name' },
  { title: 'Type', key: 'type', width: 80 },
  {
    title: 'Actions', key: 'actions', width: 160,
    render: (row: any) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'small', onClick: () => openEdit(row) }, () => 'Edit'),
        h(NButton, { size: 'small', type: 'error', onClick: () => deleteStrategy(row) }, () => 'Delete'),
      ]),
  },
]

onMounted(() => fetchStrategies())

async function fetchStrategies() {
  loading.value = true
  try {
    const res = await adminGetStrategies()
    strategies.value = res.data.data
  } catch {
    message.error('Failed to load strategies')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.name = ''
  form.type = 'local'
  form.localRoot = '/data/images'
  form.localUrl = ''
  form.s3Endpoint = ''
  form.s3Region = 'us-east-1'
  form.s3Bucket = ''
  form.s3AccessKey = ''
  form.s3SecretKey = ''
  form.s3Url = ''
  showModal.value = true
}

function openEdit(s: any) {
  editing.value = s
  form.name = s.name
  form.type = s.type
  const c = s.configs || {}
  if (s.type === 'local') {
    form.localRoot = c.root || ''
    form.localUrl = c.url_prefix || ''
  } else {
    form.s3Endpoint = c.endpoint || ''
    form.s3Region = c.region || ''
    form.s3Bucket = c.bucket || ''
    form.s3AccessKey = c.access_key || ''
    form.s3SecretKey = c.secret_key || ''
    form.s3Url = c.url_prefix || ''
  }
  showModal.value = true
}

async function saveStrategy() {
  if (!form.name) { message.warning('Name is required'); return false }
  let configs: Record<string, unknown> = {}
  if (form.type === 'local') {
    configs = { root: form.localRoot, url_prefix: form.localUrl }
  } else {
    configs = { endpoint: form.s3Endpoint, region: form.s3Region, bucket: form.s3Bucket, access_key: form.s3AccessKey, secret_key: form.s3SecretKey, url_prefix: form.s3Url }
  }
  try {
    if (editing.value) {
      await adminUpdateStrategy(editing.value.id, { name: form.name, configs })
      message.success('Strategy updated')
    } else {
      await adminCreateStrategy({ name: form.name, type: form.type, configs })
      message.success('Strategy created')
    }
    fetchStrategies()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to save strategy')
  }
  return true
}

async function deleteStrategy(s: any) {
  try {
    await adminDeleteStrategy(s.id)
    message.success('Strategy deleted')
    fetchStrategies()
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to delete strategy')
  }
}
</script>
