<template>
  <a-layout-sider
    :width="220"
    :collapsed="collapsed"
    :collapsible="true"
    :trigger="null"
    :style="{ background: '#fff', borderRight: '1px solid var(--color-border)' }"
  >
    <div class="logo">AI Gateway</div>
    <a-menu
      :selected-keys="[activeKey]"
      :style="{ borderRight: 'none' }"
    >
      <a-menu-item key="dashboard" @click="go('/')">
        <template #icon><icon-dashboard /></template>
        {{ t('sidebar.dashboard') }}
      </a-menu-item>
      <a-menu-item key="config" @click="go('/config')">
        <template #icon><icon-settings /></template>
        {{ t('sidebar.config') }}
      </a-menu-item>
    </a-menu>
  </a-layout-sider>
</template>

<script setup lang="ts">
import { IconDashboard, IconSettings } from '@arco-design/web-vue/es/icon'
import { useRouter } from 'vue-router'
import { useTranslate } from '../locale'

defineProps<{
  activeKey: string
  collapsed?: boolean
}>()

const router = useRouter()
const { t } = useTranslate()

function go(path: string) {
  router.push(path)
}
</script>

<style scoped>
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
  color: rgb(var(--primary-6));
  border-bottom: 1px solid var(--color-border);
}
</style>
