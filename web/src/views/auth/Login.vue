<template>
  <div style="display: flex; align-items: center; justify-content: center; min-height: 100vh; background-color: #f9fafb;">
    <n-card title="登录" style="width: 400px;">
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="your@email.com" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="请输入密码" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" attr-type="submit">登录</n-button>
      </n-form>
      <p style="margin-top: 16px; text-align: center; font-size: 14px; color: #6b7280;">
        还没有账号？<router-link to="/register" style="color: #3b82f6;">注册</router-link>
      </p>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { login } from '../../api/auth'
import { useUserStore } from '../../stores/user'

const router = useRouter()
const message = useMessage()
const userStore = useUserStore()
const loading = ref(false)
const formRef = ref()

const form = reactive({ email: '', password: '' })
const rules = {
  email: { required: true, message: '请输入邮箱', trigger: 'blur' },
  password: { required: true, message: '请输入密码', trigger: 'blur' },
}

async function onSubmit() {
  loading.value = true
  try {
    const res = await login(form.email, form.password)
    const { access_token, refresh_token } = res.data.data
    userStore.setTokens(access_token, refresh_token)
    await userStore.fetchProfile()
    message.success('登录成功')
    router.push('/')
  } catch (err: any) {
    message.error(err.response?.data?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>
