<template>
  <a-row :gutter="16">
    <a-col :span="8">
      <a-card class="stat-card">
        <template #title>
          <a-space>
            <icon-check-circle-fill style="color: rgb(var(--green-6))" />
            <span>{{ t('dashboard.systemStatus') }}</span>
          </a-space>
        </template>
        <div class="stat-body">
          <a-tag color="green" size="large">{{ t('dashboard.running') }}</a-tag>
          <p class="stat-desc">{{ t('dashboard.adminLabel') }}：{{ adminInfo?.username || '-' }}</p>
        </div>
      </a-card>
    </a-col>
    <a-col :span="8">
      <a-card class="stat-card">
        <template #title>
          <a-space>
            <icon-code /> <span>{{ t('dashboard.gatewayVersion') }}</span>
          </a-space>
        </template>
        <div class="stat-body">
          <div class="stat-value">{{ dashboardData?.version || '-' }}</div>
          <p class="stat-desc">{{ t('dashboard.desc') }}</p>
        </div>
      </a-card>
    </a-col>
    <a-col :span="8">
      <a-card class="stat-card">
        <template #title>
          <a-space>
            <icon-link /> <span>{{ t('dashboard.upstream') }}</span>
          </a-space>
        </template>
        <div class="stat-body">
          <div class="stat-value-url">{{ gatewayUpstream }}</div>
          <p class="stat-desc">{{ t('dashboard.clientCount') }}：{{ clientCount }}</p>
        </div>
      </a-card>
    </a-col>
  </a-row>

  <a-card :title="t('dashboard.gatewayConfig')">
    <a-descriptions v-if="gatewayConfig" :data="gatewayConfig" :column="2" />
    <a-empty v-else :description="t('dashboard.loading')" />
  </a-card>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  IconCheckCircleFill, IconCode, IconLink,
} from '@arco-design/web-vue/es/icon'
import { getDashboard } from '../api/admin'
import { useRouter } from 'vue-router'
import { useTranslate } from '../locale'

const router = useRouter()
const { t } = useTranslate()
const dashboardData = ref<any>(null)

const adminInfo = computed(() => dashboardData.value?.admin || null)
const gatewayUpstream = computed(() => {
  return dashboardData.value?.gateway?.upstream_url || '-'
})
const clientCount = computed(() => 0)
const gatewayConfig = computed(() => {
  const gw = dashboardData.value?.gateway
  if (!gw) return null
  return [
    { label: t('dashboard.address'), value: gw.address },
    { label: t('dashboard.port'), value: String(gw.port) },
    { label: t('dashboard.protocol'), value: gw.protocol },
    { label: t('dashboard.upstreamUrl'), value: gw.upstream_url },
  ]
})

onMounted(async () => {
  try {
    dashboardData.value = await getDashboard()
  } catch {
    router.push('/login')
  }
})
</script>

<style scoped>
.stat-card {
  margin-bottom: 16px;
}

.stat-card :deep(.arco-card-header) {
  height: 44px;
  border-bottom: none;
}

.stat-body {
  padding: 4px 0;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: rgb(var(--primary-6));
  line-height: 1.4;
}

.stat-value-url {
  font-size: 14px;
  font-weight: 500;
  color: rgb(var(--primary-6));
  word-break: break-all;
  line-height: 1.4;
}

.stat-desc {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--color-text-3);
}
</style>
