import { createRouter, createWebHistory } from 'vue-router'
import { getStatus, getToken } from '../api/admin'
import Init from '../views/Init.vue'
import Login from '../views/Login.vue'
import AppLayout from '../components/AppLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/init',
      name: 'init',
      component: Init,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          meta: { title: '仪表盘' },
          component: () => import('../views/Dashboard.vue'),
        },
        {
          path: 'config',
          name: 'config',
          meta: { title: '配置管理' },
          component: () => import('../views/ConfigManagement.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const publicPages = ['/init', '/login']

  try {
    const status = await getStatus()

    if (!status.initialized && to.path !== '/init') {
      return { name: 'init' }
    }

    if (status.initialized && to.path === '/init') {
      return { name: 'login' }
    }

    if (to.meta.requiresAuth && !getToken()) {
      return { name: 'login' }
    }

    if (status.initialized && to.path === '/login' && getToken()) {
      return { name: 'dashboard' }
    }
  } catch {
    if (!publicPages.includes(to.path)) {
      return { name: 'init' }
    }
  }
})

export default router
