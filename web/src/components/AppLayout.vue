<template>
  <div class="app-layout">
    <a-layout :style="{ minHeight: '100vh' }">
      <AppSidebar :activeKey="activeKey" />

      <a-layout>
        <AppHeader
          :pageTitle="pageTitle"
          :adminName="appState.adminName"
          @toggle-sidebar="sidebarCollapsed = !sidebarCollapsed"
        />

        <a-layout-content class="page-content">
          <router-view />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { appState } from '../stores/app'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'

const route = useRoute()
const sidebarCollapsed = ref(false)

const activeKey = computed(() => (route.name as string) || 'dashboard')
const pageTitle = computed(() => (route.meta?.title as string) || 'AI Gateway')
</script>

<style scoped>
.app-layout {
  height: 100vh;
  overflow: hidden;
}

.page-content {
  padding: 24px;
  height: calc(100vh - 64px);
  overflow-y: auto;
}
</style>
