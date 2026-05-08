<template>
  <a-spin :loading="loading" :tip="t('config.loading')">
    <a-tabs v-model:activeKey="activeTab" type="rounded">
      <a-tab-pane key="server" :title="t('config.tabs.server')">
        <a-form :model="cfg.server" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="8">
              <a-form-item :label="t('config.server.port')" field="server.port">
                <a-input-number v-model="cfg.server.port" :min="1" :max="65535" style="width:100%" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-divider>{{ t('config.server.tls') }}</a-divider>
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item :label="t('config.server.cert')" field="server.tls.cert">
                <a-input v-model="cfg.server.tls.cert" :placeholder="t('config.server.certPlaceholder')" />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item :label="t('config.server.key')" field="server.tls.key">
                <a-input v-model="cfg.server.tls.key" :placeholder="t('config.server.keyPlaceholder')" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="upstream" :title="t('config.tabs.upstream')">
        <a-form :model="cfg.upstream" layout="vertical">
          <a-form-item :label="t('config.upstream.url')" field="upstream.url">
            <a-input v-model="cfg.upstream.url" :placeholder="t('config.upstream.placeholder')" />
          </a-form-item>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="oauth" :title="t('config.tabs.oauth')">
        <a-alert type="info" class="mb-16">
          {{ t('config.oauth.alert') }}
        </a-alert>
        <a-form :model="cfg.oauth" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="24">
              <a-form-item :label="t('config.oauth.accessToken')" field="oauth.access_token">
                <a-input-password v-model="cfg.oauth.access_token" :placeholder="t('config.oauth.accessToken')" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-row :gutter="16">
            <a-col :span="18">
              <a-form-item :label="t('config.oauth.refreshToken')" field="oauth.refresh_token">
                <a-input-password v-model="cfg.oauth.refresh_token" :placeholder="t('config.oauth.refreshToken')" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.oauth.expiresAt')" field="oauth.expires_at">
                <a-input-number v-model="cfg.oauth.expires_at" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="auth" :title="t('config.tabs.auth')">
        <a-alert type="info" class="mb-16">
          {{ t('config.auth.alert') }}
        </a-alert>
        <a-table :data="cfg.auth.tokens" :pagination="false" class="mb-16">
          <template #columns>
            <a-table-column :title="t('config.auth.name')" data-index="name">
              <template #cell="{ record }">
                <a-input v-model="record.name" :placeholder="t('config.auth.namePlaceholder')" />
              </template>
            </a-table-column>
            <a-table-column :title="t('config.auth.token')" data-index="token">
              <template #cell="{ record }">
                <a-input-password v-model="record.token" :placeholder="t('config.auth.tokenPlaceholder')" />
              </template>
            </a-table-column>
            <a-table-column :title="t('config.auth.action')" :width="100">
              <template #cell="{ rowIndex }">
                <a-button type="text" status="danger" @click="removeToken(rowIndex)">{{ t('config.auth.delete') }}</a-button>
              </template>
            </a-table-column>
          </template>
        </a-table>
        <a-button @click="addToken">{{ t('config.auth.add') }}</a-button>
      </a-tab-pane>

      <a-tab-pane key="identity" :title="t('config.tabs.identity')">
        <a-alert type="warning" class="mb-16">
          {{ t('config.identity.alert') }}
        </a-alert>
        <a-form :model="cfg.identity" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="18">
              <a-form-item :label="t('config.identity.deviceId')" field="identity.device_id">
                <a-input v-model="cfg.identity.device_id" :placeholder="t('config.identity.deviceIdPlaceholder')" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.identity.email')" field="identity.email">
                <a-input v-model="cfg.identity.email" :placeholder="t('config.identity.emailPlaceholder')" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="env" :title="t('config.tabs.env')">
        <a-alert type="info" class="mb-16">
          {{ t('config.env.alert') }}
        </a-alert>
        <a-form :model="cfg.env" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="6"><a-form-item :label="t('config.env.platform')"><a-input v-model="cfg.env.platform" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.platformRaw')"><a-input v-model="cfg.env.platform_raw" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.arch')"><a-input v-model="cfg.env.arch" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.nodeVersion')"><a-input v-model="cfg.env.node_version" /></a-form-item></a-col>
          </a-row>
          <a-row :gutter="16">
            <a-col :span="6"><a-form-item :label="t('config.env.terminal')"><a-input v-model="cfg.env.terminal" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.packageManagers')"><a-input v-model="cfg.env.package_managers" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.runtimes')"><a-input v-model="cfg.env.runtimes" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.version')"><a-input v-model="cfg.env.version" /></a-form-item></a-col>
          </a-row>
          <a-row :gutter="16">
            <a-col :span="6"><a-form-item :label="t('config.env.versionBase')"><a-input v-model="cfg.env.version_base" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.buildTime')"><a-input v-model="cfg.env.build_time" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.deploymentEnv')"><a-input v-model="cfg.env.deployment_environment" /></a-form-item></a-col>
            <a-col :span="6"><a-form-item :label="t('config.env.vcs')"><a-input v-model="cfg.env.vcs" /></a-form-item></a-col>
          </a-row>
          <a-row :gutter="16">
            <a-col :span="4">
              <a-form-item :label="t('config.env.runningWithBun')">
                <a-switch v-model="cfg.env.is_running_with_bun" />
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item :label="t('config.env.isCi')">
                <a-switch v-model="cfg.env.is_ci" />
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item :label="t('config.env.claudeAiAuth')">
                <a-switch v-model="cfg.env.is_claude_ai_auth" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="prompt_env" :title="t('config.tabs.promptEnv')">
        <a-alert type="info" class="mb-16">
          <span v-html="t('config.promptEnv.alert')" />
        </a-alert>
        <a-form :model="cfg.prompt_env" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="6">
              <a-form-item :label="t('config.promptEnv.platform')" field="prompt_env.platform">
                <a-input v-model="cfg.prompt_env.platform" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.promptEnv.shell')" field="prompt_env.shell">
                <a-input v-model="cfg.prompt_env.shell" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.promptEnv.osVersion')" field="prompt_env.os_version">
                <a-input v-model="cfg.prompt_env.os_version" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.promptEnv.workingDir')" field="prompt_env.working_dir">
                <a-input v-model="cfg.prompt_env.working_dir" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="process" :title="t('config.tabs.process')">
        <a-alert type="info" class="mb-16">
          {{ t('config.process.alert') }}
        </a-alert>
        <a-form :model="cfg.process" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="8">
              <a-form-item :label="t('config.process.constrainedMemory')" field="process.constrained_memory">
                <a-input-number v-model="cfg.process.constrained_memory" :min="0" :step="1048576" style="width:100%" />
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item :label="t('config.process.rssMin')">
                <a-input-number v-model="cfg.process.rss_range[0]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item :label="t('config.process.rssMax')">
                <a-input-number v-model="cfg.process.rss_range[1]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
          </a-row>
          <a-row :gutter="16">
            <a-col :span="6">
              <a-form-item :label="t('config.process.heapTotalMin')">
                <a-input-number v-model="cfg.process.heap_total_range[0]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.process.heapTotalMax')">
                <a-input-number v-model="cfg.process.heap_total_range[1]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.process.heapUsedMin')">
                <a-input-number v-model="cfg.process.heap_used_range[0]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item :label="t('config.process.heapUsedMax')">
                <a-input-number v-model="cfg.process.heap_used_range[1]" :min="0" style="width:100%" />
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="logging" :title="t('config.tabs.logging')">
        <a-form :model="cfg.logging" layout="vertical">
          <a-row :gutter="16">
            <a-col :span="8">
              <a-form-item :label="t('config.logging.level')" field="logging.level">
                <a-select v-model="cfg.logging.level">
                  <a-option value="debug">Debug</a-option>
                  <a-option value="info">Info</a-option>
                  <a-option value="warn">Warn</a-option>
                  <a-option value="error">Error</a-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="8">
              <a-form-item :label="t('config.logging.audit')">
                <a-space>
                  <a-switch v-model="cfg.logging.audit" />
                  <span>{{ t('config.logging.auditDesc') }}</span>
                </a-space>
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </a-tab-pane>
    </a-tabs>

    <div class="config-actions">
      <a-space>
        <a-button @click="reloadConfig" :loading="reloading">{{ t('config.reload') }}</a-button>
        <a-button type="primary" @click="saveConfig" :loading="saving">{{ t('config.save') }}</a-button>
      </a-space>
    </div>
  </a-spin>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { getConfig, updateConfig, reloadConfig as apiReloadConfig } from '../api/admin'
import type { FullConfig } from '../api/admin'
import { useTranslate } from '../locale'

const { t } = useTranslate()
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
    Message.warning(t('config.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await updateConfig(cfg)
    Message.success(t('config.saved'))
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || t('config.saveFailed')
    Message.error(msg)
  } finally {
    saving.value = false
  }
}

async function reloadConfig() {
  reloading.value = true
  try {
    await apiReloadConfig()
    Message.success(t('config.reloaded'))
  } catch (err: any) {
    const msg = err?.response?.data?.error || err?.message || t('config.reloadFailed')
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

onMounted(loadConfig)
</script>

<style scoped>
.config-actions {
  margin-top: 24px;
  text-align: right;
}

.mb-16 {
  margin-bottom: 16px;
}
</style>
