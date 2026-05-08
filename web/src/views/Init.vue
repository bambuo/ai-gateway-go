<template>
  <div class="init-container">
    <div class="init-card">
      <div class="init-header">
        <h2>系统初始化</h2>
        <p class="init-desc">首次使用需要配置网关参数并创建管理员账户</p>
      </div>

      <a-form
        :model="form"
        :rules="rules"
        layout="vertical"
        ref="formRef"
        @submit="handleSubmit"
      >
        <a-divider>网关配置</a-divider>

        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item field="gateway.address" label="网关地址">
              <a-input
                v-model="form.gateway.address"
                placeholder="例如: 0.0.0.0 或 192.168.1.1"
                allow-clear
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="gateway.port" label="网关端口">
              <a-input-number
                v-model="form.gateway.port"
                :min="1"
                :max="65535"
                placeholder="例如: 8443"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item field="gateway.protocol" label="协议">
              <a-select
                v-model="form.gateway.protocol"
                placeholder="选择协议"
              >
                <a-option value="http">HTTP</a-option>
                <a-option value="https">HTTPS</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="24">
            <a-form-item field="gateway.upstream_url" label="上游地址">
              <a-input
                v-model="form.gateway.upstream_url"
                placeholder="默认: https://api.anthropic.com"
                allow-clear
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider>管理员账户</a-divider>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="admin.username" label="用户名">
              <a-input
                v-model="form.admin.username"
                placeholder="3-64 个字符"
                allow-clear
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="admin.email" label="邮箱">
              <a-input
                v-model="form.admin.email"
                placeholder="admin@example.com"
                allow-clear
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item field="admin.password" label="密码">
              <a-input-password
                v-model="form.admin.password"
                placeholder="至少 8 位，含大小写字母和数字"
              />
              <template #extra>
                <a-progress
                  :percent="passwordStrength"
                  :color="passwordColor"
                  :style="{ width: '100%', marginTop: '4px' }"
                  :show-text="false"
                  size="small"
                />
                <span v-if="passwordStrength > 0" :style="{ fontSize: '12px', color: passwordColor }">
                  {{ passwordLabel }}
                </span>
              </template>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item field="admin.confirm_password" label="确认密码">
              <a-input-password
                v-model="form.admin.confirm_password"
                placeholder="再次输入密码"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            :loading="submitting"
            long
            size="large"
          >
            {{ submitting ? '初始化中...' : '初始化系统' }}
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Message } from '@arco-design/web-vue'
import { initSystem } from '../api/admin'
import { useRouter } from 'vue-router'

const router = useRouter()
const formRef = ref<any>(null)
const submitting = ref(false)

const form = ref({
  gateway: {
    address: '',
    port: 8443,
    protocol: 'http',
    upstream_url: 'https://api.anthropic.com',
  },
  admin: {
    username: '',
    password: '',
    confirm_password: '',
    email: '',
  },
})

const passwordStrength = computed(() => {
  const pwd = form.value.admin.password
  if (!pwd) return 0
  let score = 0
  if (pwd.length >= 8) score += 25
  if (pwd.length >= 12) score += 10
  if (/[a-z]/.test(pwd)) score += 15
  if (/[A-Z]/.test(pwd)) score += 15
  if (/[0-9]/.test(pwd)) score += 15
  if (/[^a-zA-Z0-9]/.test(pwd)) score += 20
  return Math.min(score, 100)
})

const passwordColor = computed(() => {
  if (passwordStrength.value < 40) return '#f53f3f'
  if (passwordStrength.value < 70) return '#ff7d00'
  if (passwordStrength.value < 90) return '#ffb400'
  return '#00b42a'
})

const passwordLabel = computed(() => {
  if (passwordStrength.value === 0) return ''
  if (passwordStrength.value < 40) return '弱'
  if (passwordStrength.value < 70) return '中'
  if (passwordStrength.value < 90) return '强'
  return '非常强'
})

const rules = {
  'gateway.address': [{ required: true, message: '请输入网关地址' }],
  'gateway.port': [
    { required: true, message: '请输入网关端口' },
    {
      validator: (value: any, callback: any) => {
        if (value < 1 || value > 65535) {
          callback('端口必须在 1-65535 之间')
        } else {
          callback()
        }
      },
    },
  ],
  'gateway.protocol': [{ required: true, message: '请选择协议' }],
  'gateway.upstream_url': [{ required: true, message: '请输入上游地址' }],
  'admin.username': [
    { required: true, message: '请输入用户名' },
    {
      validator: (value: any, callback: any) => {
        if (value && (value.length < 3 || value.length > 64)) {
          callback('用户名长度必须在 3-64 个字符之间')
        } else {
          callback()
        }
      },
    },
  ],
  'admin.password': [
    { required: true, message: '请输入密码' },
    {
      validator: (value: any, callback: any) => {
        if (!value) return callback()
        if (value.length < 8) return callback('密码长度不能少于 8 位')
        if (!/[a-z]/.test(value)) return callback('密码需要包含小写字母')
        if (!/[A-Z]/.test(value)) return callback('密码需要包含大写字母')
        if (!/[0-9]/.test(value)) return callback('密码需要包含数字')
        callback()
      },
    },
  ],
  'admin.confirm_password': [
    { required: true, message: '请再次输入密码' },
    {
      validator: (value: any, callback: any) => {
        if (value !== form.value.admin.password) {
          callback('两次输入的密码不一致')
        } else {
          callback()
        }
      },
    },
  ],
  'admin.email': [
    { required: true, message: '请输入邮箱' },
    {
      validator: (_value: any, callback: any) => {
        const email = form.value.admin.email
        if (!email) return callback()
        if (!/^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/.test(email)) {
          callback('邮箱格式不正确')
        } else {
          callback()
        }
      },
    },
  ],
}

async function handleSubmit() {
  submitting.value = true
  try {
    await initSystem(form.value)
    Message.success('系统初始化成功！即将跳转至仪表盘...')
    setTimeout(() => router.push('/'), 1500)
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '初始化失败，请重试'
    Message.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.init-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 24px;
}

.init-card {
  width: 100%;
  max-width: 680px;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

.init-header {
  text-align: center;
  margin-bottom: 8px;
}

.init-header h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: var(--color-text-1);
}

.init-desc {
  margin: 0;
  color: var(--color-text-3);
  font-size: 14px;
}
</style>
