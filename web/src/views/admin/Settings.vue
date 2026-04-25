<template>
  <div>
    <n-h3 style="margin-bottom: 16px;">系统设置</n-h3>
    <n-spin :show="loading">
      <n-form v-if="settings" label-placement="left" label-width="120" style="max-width: 600px;">
        <n-form-item label="站点名称">
          <n-input v-model:value="settings.app_name" />
        </n-form-item>
        <n-form-item label="站点地址">
          <n-input v-model:value="settings.app_url" placeholder="https://your-domain.com" />
        </n-form-item>
        <n-form-item label="开放注册">
          <n-switch v-model:value="settings.allow_registration" />
        </n-form-item>
        <n-form-item label="游客上传">
          <n-switch v-model:value="settings.allow_guest_upload" />
        </n-form-item>
        <n-form-item>
          <n-button type="primary" :loading="saving" @click="saveSettings">保存设置</n-button>
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
    message.error('加载设置失败')
  } finally {
    loading.value = false
  }
})

async function saveSettings() {
  if (!settings.value) return
  saving.value = true
  try {
    await adminUpdateSettings(settings.value)
    message.success('设置已保存')
  } catch (err: any) {
    message.error(err.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
