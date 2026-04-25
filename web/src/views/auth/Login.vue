<template>
  <div class="flex items-center justify-center min-h-screen bg-gray-50">
    <n-card title="Login" class="w-96">
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
        <n-form-item label="Email" path="email">
          <n-input v-model:value="form.email" placeholder="your@email.com" />
        </n-form-item>
        <n-form-item label="Password" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="Password" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" attr-type="submit">Login</n-button>
      </n-form>
      <p class="mt-4 text-center text-sm text-gray-500">
        Don't have an account? <router-link to="/register" class="text-blue-500">Register</router-link>
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
  email: { required: true, message: 'Email is required', trigger: 'blur' },
  password: { required: true, message: 'Password is required', trigger: 'blur' },
}

async function onSubmit() {
  loading.value = true
  try {
    const res = await login(form.email, form.password)
    const { access_token, refresh_token } = res.data.data
    userStore.setTokens(access_token, refresh_token)
    await userStore.fetchProfile()
    message.success('Login successful')
    router.push('/')
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Login failed')
  } finally {
    loading.value = false
  }
}
</script>
