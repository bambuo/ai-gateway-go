import { createRouter, createWebHistory } from 'vue-router'
import { getStatus, getToken } from '../api/admin'
import Init from '../views/Init.vue'
import Login from '../views/Login.vue'

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
      name: 'dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true },
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
