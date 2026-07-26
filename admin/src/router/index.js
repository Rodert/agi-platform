import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layout/Index.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据统计', icon: 'DataAnalysis' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', icon: 'User' }
      },
      {
        path: 'user-defaults',
        name: 'UserDefaults',
        component: () => import('@/views/UserDefaults.vue'),
        meta: { title: '用户默认设置', icon: 'User' }
      },
      {
        path: 'redeem-codes',
        name: 'RedeemCodes',
        component: () => import('@/views/RedeemCodes.vue'),
        meta: { title: '兑换码管理', icon: 'Ticket' }
      },
      {
        path: 'works',
        name: 'Works',
        component: () => import('@/views/Works.vue'),
        meta: { title: '作品审核', icon: 'Picture' }
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('@/views/Tasks.vue'),
        meta: { title: '生成任务', icon: 'List' }
      },
      {
        path: 'announcements',
        name: 'Announcements',
        component: () => import('@/views/Announcements.vue'),
        meta: { title: '全员通知', icon: 'Bell' }
      },
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('@/views/Reports.vue'),
        meta: { title: '数据报表', icon: 'TrendCharts' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/Logs.vue'),
        meta: { title: '日志管理', icon: 'Document' }
      },
      {
        path: 'database',
        name: 'Database',
        component: () => import('@/views/Database.vue'),
        meta: { title: '数据管理', icon: 'Grid' }
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: () => import('@/views/Permissions.vue'),
        meta: { title: '管理员与权限', icon: 'Lock' }
      },
      {
        path: 'accounts',
        name: 'Accounts',
        component: () => import('@/views/Accounts.vue'),
        meta: { title: '渠道与模型', icon: 'Key' }
      },
      {
        path: 'config',
        redirect: '/platform-settings'
      },
      {
        path: 'platform-settings',
        name: 'PlatformSettings',
        component: () => import('@/views/Config.vue'),
        meta: { title: '站点与邮件', configTabs: ['basic', 'email'] }
      },
      {
        path: 'credit-packages',
        name: 'CreditPackages',
        component: () => import('@/views/Config.vue'),
        meta: { title: '充值套餐', configTabs: ['packages'] }
      },
      {
        path: 'task-policy',
        name: 'TaskPolicy',
        component: () => import('@/views/Config.vue'),
        meta: { title: '任务策略', configTabs: ['task'] }
      },
      {
        path: 'prompt-optimization',
        name: 'PromptOptimization',
        component: () => import('@/views/Config.vue'),
        meta: { title: '提示词优化', configTabs: ['prompt-optimization'] }
      },
      {
        path: 'storage',
        name: 'Storage',
        component: () => import('@/views/Config.vue'),
        meta: { title: '存储与资源', configTabs: ['storage', 'resources'] }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { title: '个人中心' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.token) {
    next('/login')
  } else if (to.path === '/login' && authStore.token) {
    next('/')
  } else {
    next()
  }
})

export default router
