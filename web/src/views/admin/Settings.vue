<template>
  <div>
    <n-h3 class="mb-4">Settings</n-h3>
    <n-spin :show="loading">
      <n-form v-if="settings" label-placement="left" label-width="200" class="max-w-2xl">
        <n-form-item label="App Name">
          <n-input v-model:value="settings.app_name" />
        </n-form-item>
        <n-form-item label="App URL">
          <n-input v-model:value="settings.app_url" placeholder="https://your-domain.com" />
        </n-form-item>
        <n-form-item label="Allow Registration">
          <n-switch v-model:value="settings.allow_registration" />
        </n-form-item>
        <n-form-item label="Allow Guest Upload">
          <n-switch v-model:value="settings.allow_guest_upload" />
        </n-form-item>
        <n-form-item>
          <n-button type="primary" :loading="saving" @click="saveSettings">Save Settings</n-button>
        </n-form-item>
      </n-form>
    </n-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NH3, NForm, NFormItem, NInput, NSwitch, NButton, NSpin, useMessage } from 'naive-ui'
import { adminGetSettings, adminUpdateSettings } from '../../api/admin'

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const settings = ref<Record<string, any> | null>(null)

onMounted(async () => {
  loading.value = true
  try {
    const res = await adminGetSettings()
    settings.value = res.data.data
  } catch {
    message.error('Failed to load settings')
  } finally {
    loading.value = false
  }
})

async function saveSettings() {
  if (!settings.value) return
  saving.value = true
  try {
    await adminUpdateSettings(settings.value)
    message.success('Settings saved')
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Failed to save settings')
  } finally {
    saving.value = false
  }
}
</script>
