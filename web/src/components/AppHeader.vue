<template>
  <header class="app-header" role="banner">
    <div class="header-left">
      <button
        class="menu-toggle"
        :aria-label="t('header.menuToggle')"
        aria-expanded="false"
        @click="$emit('toggle-sidebar')"
      >
        <icon-menu-fold v-if="!collapsed" />
        <icon-menu-unfold v-else />
      </button>
      <a-breadcrumb>
        <a-breadcrumb-item>
          <icon-home />
        </a-breadcrumb-item>
        <a-breadcrumb-item>{{ pageTitle }}</a-breadcrumb-item>
      </a-breadcrumb>
    </div>

    <div class="header-right" :class="{ 'mobile-open': mobileOpen }">
      <button
        class="icon-btn"
        :aria-label="t('header.themeLight')"
        :title="t('header.themeLight')"
        @click="toggleTheme"
      >
        <icon-sun v-if="appState.theme === 'light'" />
        <icon-moon v-else />
      </button>

      <div class="lang-select" role="radiogroup" :aria-label="t('header.language')">
        <button
          class="lang-btn"
          :class="{ active: appState.lang === 'zh-CN' }"
          :aria-label="'中文'"
          :aria-checked="appState.lang === 'zh-CN'"
          role="radio"
          @click="setLang('zh-CN')"
        >
          中
        </button>
        <button
          class="lang-btn"
          :class="{ active: appState.lang === 'en-US' }"
          :aria-label="'English'"
          :aria-checked="appState.lang === 'en-US'"
          role="radio"
          @click="setLang('en-US')"
        >
          EN
        </button>
      </div>

      <a-trigger trigger="click" position="br" :showDelay="0">
        <button class="user-btn" aria-label="User menu" aria-haspopup="true">
          <a-avatar :size="32" :style="{ background: 'rgb(var(--primary-6))', cursor: 'pointer' }">
            {{ adminName?.charAt(0)?.toUpperCase() || 'A' }}
          </a-avatar>
          <span class="user-name">{{ adminName || '-' }}</span>
        </button>
        <template #content>
          <a-menu @menu-item-click="handleMenuClick">
            <a-menu-item key="profile">
              <template #icon><icon-user /></template>
              {{ t('header.profile') }}
            </a-menu-item>
            <a-menu-item key="settings">
              <template #icon><icon-settings /></template>
              {{ t('header.accountSettings') }}
            </a-menu-item>
            <a-menu-item key="logout">
              <template #icon><icon-export /></template>
              {{ t('header.logout') }}
            </a-menu-item>
          </a-menu>
        </template>
      </a-trigger>
    </div>

    <button
      class="mobile-toggle"
      :aria-label="t('header.menuToggle')"
      aria-expanded="false"
      @click="mobileOpen = !mobileOpen"
    >
      <icon-menu-fold v-if="!mobileOpen" />
      <icon-menu-unfold v-else />
    </button>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  IconMenuFold, IconMenuUnfold, IconHome,
  IconSun, IconMoon,
  IconUser, IconSettings, IconExport,
} from '@arco-design/web-vue/es/icon'
import { logout } from '../api/admin'
import { appState, toggleTheme, setLang } from '../stores/app'
import { useTranslate } from '../locale'

defineProps<{
  pageTitle: string
  adminName?: string
  collapsed?: boolean
}>()

defineEmits<{
  'toggle-sidebar': []
}>()

const router = useRouter()
const { t } = useTranslate()
const mobileOpen = ref(false)

function handleMenuClick(key: string | number) {
  if (key === 'logout') {
    logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.app-header {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.menu-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  color: var(--color-text-2);
  font-size: 18px;
  transition: background 0.2s, color 0.2s;
}

.menu-toggle:hover {
  background: var(--color-fill-2);
  color: var(--color-text-1);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  color: var(--color-text-2);
  font-size: 18px;
  transition: background 0.2s, color 0.2s;
}

.icon-btn:hover {
  background: var(--color-fill-2);
  color: var(--color-text-1);
}

.lang-select {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  background: var(--color-fill-2);
  border-radius: 6px;
}

.lang-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 26px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-3);
  transition: all 0.2s;
}

.lang-btn.active {
  background: #fff;
  color: rgb(var(--primary-6));
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

body[arco-theme='dark'] .lang-btn.active {
  background: var(--color-bg-2);
}

.lang-btn:hover {
  color: var(--color-text-1);
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 4px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.user-btn:hover {
  background: var(--color-fill-2);
}

.user-name {
  font-size: 14px;
  color: var(--color-text-1);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mobile-toggle {
  display: none;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  color: var(--color-text-2);
  font-size: 18px;
}

@media (max-width: 768px) {
  .header-right {
    display: none;
    position: absolute;
    top: 64px;
    left: 0;
    right: 0;
    background: #fff;
    border-bottom: 1px solid var(--color-border);
    padding: 12px 16px;
    flex-direction: row-reverse;
    flex-wrap: wrap;
    justify-content: flex-end;
    z-index: 100;
  }

  .header-right.mobile-open {
    display: flex;
  }

  .mobile-toggle {
    display: flex;
  }

  .menu-toggle {
    display: none;
  }
}
</style>
