<template>
  <div class="dashboard">
    <a-layout>
      <a-layout-sider
        :width="220"
        :style="{ background: '#fff', borderRight: '1px solid var(--color-border)' }"
      >
        <div class="logo">AI Gateway</div>
        <a-menu
          :selected-keys="['dashboard']"
          :style="{ borderRight: 'none' }"
        >
          <a-menu-item key="dashboard">
            <IconDashboard /> 仪表盘
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
            <span class="header-title">仪表盘</span>
            <span v-if="adminInfo" class="header-user">
              {{ adminInfo.username }} ({{ adminInfo.email }})
            </span>
          </div>
        </a-layout-header>

        <a-layout-content :style="{ padding: '24px', minHeight: 'calc(100vh - 60px)' }">
          <a-row :gutter="16">
            <a-col :span="colSpan">
              <a-card :style="{ marginBottom: '16px' }">
                <a-statistic title="系统状态" :value="1" />
                <p>系统正常运行中</p>
              </a-card>
            </a-col>
            <a-col :span="colSpan">
              <a-card :style="{ marginBottom: '16px' }">
                <a-statistic title="版本" :value="dashboardData?.version || '-'" />
                <p>AI Gateway 管理后台</p>
              </a-card>
            </a-col>
          </a-row>

          <a-card title="网关配置" :style="{ marginBottom: '16px' }">
            <a-descriptions v-if="gatewayConfig" :data="gatewayConfig" :column="2" />
            <a-empty v-else description="加载中..." />
          </a-card>
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { IconDashboard, IconExport } from '@arco-design/web-vue/es/icon'
import { getDashboard, logout } from '../api/admin'
import { useRouter } from 'vue-router'

const router = useRouter()
const colSpan = 8
const dashboardData = ref<any>(null)

const adminInfo = computed(() => dashboardData.value?.admin || null)
const gatewayConfig = computed(() => {
  const gw = dashboardData.value?.gateway
  if (!gw) return null
  return [
    { label: '地址', value: gw.address },
    { label: '端口', value: String(gw.port) },
    { label: '协议', value: gw.protocol },
    { label: '上游地址', value: gw.upstream_url },
  ]
})

function handleLogout() {
  logout()
  router.push('/login')
}

onMounted(async () => {
  try {
    dashboardData.value = await getDashboard()
  } catch {
    // Redirect to login on failure
    router.push('/login')
  }
})
</script>

<style scoped>
.dashboard {
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

.header-user {
  font-size: 14px;
  color: var(--color-text-3);
}
</style>
