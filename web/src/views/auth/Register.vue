<template>
  <div style="display: flex; align-items: center; justify-content: center; min-height: 100vh; background-color: #f9fafb;">
    <n-card title="注册" style="width: 400px;">
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
        <n-form-item label="昵称" path="name">
          <n-input v-model:value="form.name" placeholder="你的昵称" />
        </n-form-item>
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="your@email.com" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="至少 8 位" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" attr-type="submit">注册</n-button>
      </n-form>
      <p style="margin-top: 16px; text-align: center; font-size: 14px; color: #6b7280;">
        已有账号？<router-link to="/login" style="color: #3b82f6;">登录</router-link>
      </p>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { register } from '../../api/auth'
import { useUserStore } from '../../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const formRef = ref()

const form = reactive({ name: '', email: '', password: '' })
const rules = {
  name: { required: true, message: '请输入昵称', trigger: 'blur' },
  email: { required: true, message: '请输入邮箱', trigger: 'blur' },
  password: { required: true, min: 8, message: '密码至少 8 位', trigger: 'blur' },
}

async function onSubmit() {
  loading.value = true
  try {
    const res = await register(form.email, form.password, form.name)
    const { access_token, refresh_token } = res.data.data
    userStore.setTokens(access_token, refresh_token)
    await userStore.fetchProfile()
    message.success('注册成功')
    router.push('/')
  } catch (err: any) {
    message.error(err.response?.data?.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>
