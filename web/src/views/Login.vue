<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h2>AI Gateway</h2>
        <p class="login-desc">管理后台登录</p>
      </div>

      <a-form
        :model="form"
        :rules="rules"
        layout="vertical"
        ref="formRef"
        @submit="handleLogin"
      >
        <a-form-item field="username" label="用户名">
          <a-input
            v-model="form.username"
            placeholder="请输入用户名"
            size="large"
            allow-clear
          />
        </a-form-item>

        <a-form-item field="password" label="密码">
          <a-input-password
            v-model="form.password"
            placeholder="请输入密码"
            size="large"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            :loading="submitting"
            long
            size="large"
          >
            {{ submitting ? '登录中...' : '登 录' }}
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { login } from '../api/admin'
import { useRouter } from 'vue-router'

const router = useRouter()
const formRef = ref<any>(null)
const submitting = ref(false)

const form = ref({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名' }],
  password: [{ required: true, message: '请输入密码' }],
}

async function handleLogin() {
  submitting.value = true
  try {
    const res = await login(form.value)
    Message.success(`欢迎回来，${res.admin.username}`)
    router.push('/')
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '登录失败'
    Message.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--color-bg-2);
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: var(--color-text-1);
}

.login-desc {
  margin: 0;
  color: var(--color-text-3);
  font-size: 14px;
}
</style>
