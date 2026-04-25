<template>
  <div class="flex items-center justify-center min-h-screen bg-gray-50">
    <n-card title="Register" class="w-96">
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="onSubmit">
        <n-form-item label="Name" path="name">
          <n-input v-model:value="form.name" placeholder="Your name" />
        </n-form-item>
        <n-form-item label="Email" path="email">
          <n-input v-model:value="form.email" placeholder="your@email.com" />
        </n-form-item>
        <n-form-item label="Password" path="password">
          <n-input v-model:value="form.password" type="password" placeholder="Password" />
        </n-form-item>
        <n-button type="primary" block :loading="loading" attr-type="submit">Register</n-button>
      </n-form>
      <p class="mt-4 text-center text-sm text-gray-500">
        Already have an account? <router-link to="/login" class="text-blue-500">Login</router-link>
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
  name: { required: true, message: 'Name is required', trigger: 'blur' },
  email: { required: true, message: 'Email is required', trigger: 'blur' },
  password: { required: true, min: 6, message: 'At least 6 characters', trigger: 'blur' },
}

async function onSubmit() {
  loading.value = true
  try {
    const res = await register(form.email, form.password, form.name)
    const { access_token, refresh_token } = res.data.data
    userStore.setTokens(access_token, refresh_token)
    await userStore.fetchProfile()
    message.success('Registration successful')
    router.push('/')
  } catch (err: any) {
    message.error(err.response?.data?.message || 'Registration failed')
  } finally {
    loading.value = false
  }
}
</script>
