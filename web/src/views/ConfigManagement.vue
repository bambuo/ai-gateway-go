<template>
  <div class="config-page">
    <a-layout>
      <a-layout-sider
        :width="220"
        :style="{ background: '#fff', borderRight: '1px solid var(--color-border)' }"
      >
        <div class="logo">AI Gateway</div>
        <a-menu
          :selected-keys="['config']"
          :style="{ borderRight: 'none' }"
        >
          <a-menu-item key="dashboard" @click="goDashboard">
            <IconDashboard /> 仪表盘
          </a-menu-item>
          <a-menu-item key="config">
            <IconSettings /> 配置管理
          </a-menu-item>
        </a-menu>
        <div class="logout-area">
          <a-button type="text" @click="handleLogout">
            <IconExport /> 退出登录
          </a-button>
        </div>
      </a-layout-sider>

      <a-layout>
        <a-layout-header :style="{ background: '#fff', padding: '0 24px', borderBottom: '1px solid var(--color-border)' }">
          <div class="header-content">
            <span class="header-title">配置管理</span>
            <a-space>
              <a-button @click="reloadConfig" :loading="reloading">重新加载</a-button>
              <a-button type="primary" @click="saveConfig" :loading="saving">保存配置</a-button>
            </a-space>
          </div>
        </a-layout-header>

        <a-layout-content :style="{ padding: '24px', minHeight: 'calc(100vh - 60px)' }">
          <a-spin :loading="loading" tip="加载配置中...">
            <a-tabs v-model:activeKey="activeTab" type="rounded">
              <a-tab-pane key="server" title="Server">
                <a-form :model="cfg.server" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-form-item label="监听端口" field="server.port">
                        <a-input-number v-model="cfg.server.port" :min="1" :max="65535" style="width:100%" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  <a-divider>TLS 配置</a-divider>
                  <a-row :gutter="16">
                    <a-col :span="12">
                      <a-form-item label="证书路径" field="server.tls.cert">
                        <a-input v-model="cfg.server.tls.cert" placeholder="留空则不启用 TLS" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="12">
                      <a-form-item label="密钥路径" field="server.tls.key">
                        <a-input v-model="cfg.server.tls.key" placeholder="留空则不启用 TLS" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="upstream" title="Upstream">
                <a-form :model="cfg.upstream" layout="vertical">
                  <a-form-item label="上游地址" field="upstream.url">
                    <a-input v-model="cfg.upstream.url" placeholder="https://api.anthropic.com" />
                  </a-form-item>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="oauth" title="OAuth">
                <a-alert type="info" class="mb-16">
                  网关集中管理 OAuth Token 生命周期。可通过 <code>scripts/quick-setup.sh</code> 自动提取或手动填写。
                </a-alert>
                <a-form :model="cfg.oauth" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="24">
                      <a-form-item label="Access Token" field="oauth.access_token">
                        <a-input-password v-model="cfg.oauth.access_token" placeholder="OAuth Access Token" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  <a-row :gutter="16">
                    <a-col :span="18">
                      <a-form-item label="Refresh Token" field="oauth.refresh_token">
                        <a-input-password v-model="cfg.oauth.refresh_token" placeholder="OAuth Refresh Token" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="过期时间戳" field="oauth.expires_at">
                        <a-input-number v-model="cfg.oauth.expires_at" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="auth" title="Auth">
                <a-alert type="info" class="mb-16">
                  客户端认证令牌列表。每个客户端使用唯一 Token 进行身份认证。
                </a-alert>
                <a-table :data="cfg.auth.tokens" :pagination="false" class="mb-16">
                  <template #columns>
                    <a-table-column title="名称" data-index="name">
                      <template #cell="{ record }">
                        <a-input v-model="record.name" placeholder="客户端名称" />
                      </template>
                    </a-table-column>
                    <a-table-column title="Token" data-index="token">
                      <template #cell="{ record }">
                        <a-input-password v-model="record.token" placeholder="认证令牌" />
                      </template>
                    </a-table-column>
                    <a-table-column title="操作" :width="100">
                      <template #cell="{ rowIndex }">
                        <a-button type="text" status="danger" @click="removeToken(rowIndex)">删除</a-button>
                      </template>
                    </a-table-column>
                  </template>
                </a-table>
                <a-button @click="addToken">+ 添加客户端</a-button>
              </a-tab-pane>

              <a-tab-pane key="identity" title="Identity">
                <a-alert type="warning" class="mb-16">
                  Device ID 必须是 64 位十六进制值，可通过 <code>gateway gen-identity</code> 生成。
                </a-alert>
                <a-form :model="cfg.identity" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="18">
                      <a-form-item label="Device ID" field="identity.device_id">
                        <a-input v-model="cfg.identity.device_id" placeholder="64 位十六进制值" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="邮箱" field="identity.email">
                        <a-input v-model="cfg.identity.email" placeholder="admin@example.com" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="env" title="Env">
                <a-alert type="info" class="mb-16">
                  规范化的环境指纹，所有客户端将统一显示为这些值。
                </a-alert>
                <a-form :model="cfg.env" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="6"><a-form-item label="Platform"><a-input v-model="cfg.env.platform" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Platform Raw"><a-input v-model="cfg.env.platform_raw" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Arch"><a-input v-model="cfg.env.arch" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Node Version"><a-input v-model="cfg.env.node_version" /></a-form-item></a-col>
                  </a-row>
                  <a-row :gutter="16">
                    <a-col :span="6"><a-form-item label="Terminal"><a-input v-model="cfg.env.terminal" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Package Managers"><a-input v-model="cfg.env.package_managers" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Runtimes"><a-input v-model="cfg.env.runtimes" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Version"><a-input v-model="cfg.env.version" /></a-form-item></a-col>
                  </a-row>
                  <a-row :gutter="16">
                    <a-col :span="6"><a-form-item label="Version Base"><a-input v-model="cfg.env.version_base" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Build Time"><a-input v-model="cfg.env.build_time" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="Deployment Env"><a-input v-model="cfg.env.deployment_environment" /></a-form-item></a-col>
                    <a-col :span="6"><a-form-item label="VCS"><a-input v-model="cfg.env.vcs" /></a-form-item></a-col>
                  </a-row>
                  <a-row :gutter="16">
                    <a-col :span="4">
                      <a-form-item label="Running with Bun">
                        <a-switch v-model="cfg.env.is_running_with_bun" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="4">
                      <a-form-item label="Is CI">
                        <a-switch v-model="cfg.env.is_ci" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="4">
                      <a-form-item label="Claude AI Auth">
                        <a-switch v-model="cfg.env.is_claude_ai_auth" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="prompt_env" title="PromptEnv">
                <a-alert type="info" class="mb-16">
                  系统提示词环境伪装值，替换提示词中的 &lt;env&gt; 块。
                </a-alert>
                <a-form :model="cfg.prompt_env" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="6">
                      <a-form-item label="Platform" field="prompt_env.platform">
                        <a-input v-model="cfg.prompt_env.platform" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="Shell" field="prompt_env.shell">
                        <a-input v-model="cfg.prompt_env.shell" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="OS Version" field="prompt_env.os_version">
                        <a-input v-model="cfg.prompt_env.os_version" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="Working Dir" field="prompt_env.working_dir">
                        <a-input v-model="cfg.prompt_env.working_dir" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="process" title="Process">
                <a-alert type="info" class="mb-16">
                  规范化的进程指标，应在合理范围内随机化。
                </a-alert>
                <a-form :model="cfg.process" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-form-item label="Constrained Memory" field="process.constrained_memory">
                        <a-input-number v-model="cfg.process.constrained_memory" :min="0" :step="1048576" style="width:100%" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="8">
                      <a-form-item label="RSS Range (min)">
                        <a-input-number v-model="cfg.process.rss_range[0]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="8">
                      <a-form-item label="RSS Range (max)">
                        <a-input-number v-model="cfg.process.rss_range[1]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  <a-row :gutter="16">
                    <a-col :span="6">
                      <a-form-item label="Heap Total (min)">
                        <a-input-number v-model="cfg.process.heap_total_range[0]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="Heap Total (max)">
                        <a-input-number v-model="cfg.process.heap_total_range[1]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="Heap Used (min)">
                        <a-input-number v-model="cfg.process.heap_used_range[0]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="Heap Used (max)">
                        <a-input-number v-model="cfg.process.heap_used_range[1]" :min="0" style="width:100%" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>

              <a-tab-pane key="logging" title="Logging">
                <a-form :model="cfg.logging" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-form-item label="日志级别" field="logging.level">
                        <a-select v-model="cfg.logging.level">
                          <a-option value="debug">Debug</a-option>
                          <a-option value="info">Info</a-option>
                          <a-option value="warn">Warn</a-option>
                          <a-option value="error">Error</a-option>
                        </a-select>
                      </a-form-item>
                    </a-col>
                    <a-col :span="8">
                      <a-form-item label="审计日志">
                        <a-space>
                          <a-switch v-model="cfg.logging.audit" />
                          <span>记录每个请求的客户端来源</span>
                        </a-space>
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-tab-pane>
            </a-tabs>
          </a-spin>
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { IconDashboard, IconSettings, IconExport } from '@arco-design/web-vue/es/icon'
import { Message } from '@arco-design/web-vue'
import { getConfig, updateConfig, reloadConfig as apiReloadConfig, logout } from '../api/admin'
import type { FullConfig } from '../api/admin'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const reloading = ref(false)
const activeTab = ref('server')

const defaultConfig: FullConfig = {
  server: { port: 8443, tls: { cert: '', key: '' } },
  upstream: { url: 'https://api.anthropic.com' },
  oauth: { access_token: '', refresh_token: '', expires_at: 0 },
  auth: { tokens: [] },
  identity: { device_id: '', email: '' },
  env: {
    platform: 'darwin', platform_raw: 'darwin', arch: 'arm64',
    node_version: 'v24.3.0', terminal: 'iTerm2.app',
    package_managers: 'npm,pnpm', runtimes: 'node',
    is_running_with_bun: false, is_ci: false, is_claude_ai_auth: true,
    version: '2.1.81', version_base: '2.1.81',
    build_time: '2026-03-20T21:26:18Z',
    deployment_environment: 'unknown-darwin', vcs: 'git',
  },
  prompt_env: { platform: 'darwin', shell: 'zsh', os_version: 'Darwin 24.4.0', working_dir: '/Users/user/projects' },
  process: {
    constrained_memory: 34359738368,
    rss_range: [300000000, 500000000],
    heap_total_range: [40000000, 80000000],
    heap_used_range: [100000000, 200000000],
  },
  logging: { level: 'info', audit: true },
}

const cfg = reactive<FullConfig>(JSON.parse(JSON.stringify(defaultConfig)))

async function loadConfig() {
  loading.value = true
  try {
    const data = await getConfig()
    Object.assign(cfg, data)
  } catch {
    Message.warning('加载配置失败，使用默认值')
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await updateConfig(cfg)
    Message.success('配置已保存')
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '保存失败'
    Message.error(msg)
  } finally {
    saving.value = false
  }
}

async function reloadConfig() {
  reloading.value = true
  try {
    await apiReloadConfig()
    Message.success('配置已重新加载')
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || '重新加载失败'
    Message.error(msg)
  } finally {
    reloading.value = false
  }
}

function addToken() {
  cfg.auth.tokens.push({ name: '', token: '' })
}

function removeToken(index: number) {
  cfg.auth.tokens.splice(index, 1)
}

function goDashboard() {
  router.push('/')
}

function handleLogout() {
  logout()
  router.push('/login')
}

onMounted(loadConfig)
</script>

<style scoped>
.config-page {
  min-height: 100vh;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
  color: rgb(var(--primary-6));
  border-bottom: 1px solid var(--color-border);
}

.logout-area {
  position: absolute;
  bottom: 16px;
  left: 0;
  right: 0;
  text-align: center;
}

.header-content {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
}

.mb-16 {
  margin-bottom: 16px;
}
</style>
