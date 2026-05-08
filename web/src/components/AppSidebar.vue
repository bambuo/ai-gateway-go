<template>
  <a-layout-sider
    :width="220"
    :collapsed="collapsed"
    :collapsed-width="64"
    class="app-sidebar"
    :style="{ background: 'var(--color-bg-2)', borderRight: '1px solid var(--color-border)' }"
  >
    <div class="logo">{{ collapsed ? 'AI' : 'AI Gateway' }}</div>
    <a-menu
      :selected-keys="[activeKey]"
      :collapse="collapsed"
    >
      <a-menu-item key="dashboard" @click="go('/')">
        <template #icon><icon-dashboard :size="collapsed ? 22 : 16" /></template>
        {{ t('sidebar.dashboard') }}
      </a-menu-item>
      <a-menu-item key="config" @click="go('/config')">
        <template #icon><icon-settings :size="collapsed ? 22 : 16" /></template>
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

.app-sidebar :deep(.arco-menu-collapsed) {
  width: 64px;
}

.app-sidebar :deep(.arco-menu-collapsed .arco-menu-item) {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 !important;
  width: 40px !important;
  height: 40px !important;
  margin: 8px auto;
  border-radius: 8px;
}

.app-sidebar :deep(.arco-menu-collapsed .arco-menu-item .arco-menu-icon) {
  margin-right: 0 !important;
  line-height: 1;
}

.app-sidebar :deep(.arco-menu-collapsed .arco-menu-item .arco-menu-title) {
  display: none;
}

.app-sidebar :deep(.arco-menu-collapsed .arco-menu-inner) {
  padding: 8px 0 !important;
}

.app-sidebar :deep(.arco-menu:not(.arco-menu-collapsed) .arco-menu-item) {
  display: flex;
  align-items: center;
  margin: 4px 12px;
  width: auto !important;
  border-radius: 8px;
}
</style>
