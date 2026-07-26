<template>
  <el-container class="layout">
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <h3>AGI Platform</h3>
      </div>
      <el-menu
        :default-active="activeMenu"
        :default-openeds="openMenus"
        :router="true"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-sub-menu index="creation">
          <template #title><el-icon><Picture /></el-icon><span>创作运营</span></template>
          <el-menu-item index="/tasks">生成任务</el-menu-item>
          <el-menu-item index="/works">作品与审核</el-menu-item>
          <el-menu-item index="/announcements">全员通知</el-menu-item>
          <el-menu-item index="/reports">数据报表</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="users">
          <template #title><el-icon><User /></el-icon><span>用户与权益</span></template>
          <el-menu-item index="/users">用户管理</el-menu-item>
          <el-menu-item index="/user-defaults">用户默认设置</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="credits">
          <template #title><el-icon><Coin /></el-icon><span>订单与灵感值</span></template>
          <el-menu-item index="/credit-packages">充值套餐</el-menu-item>
		  <el-menu-item index="/redeem-codes">兑换码管理</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="ai">
          <template #title><el-icon><Connection /></el-icon><span>AI 渠道与模型</span></template>
          <el-menu-item index="/accounts">渠道与模型</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="generation">
          <template #title><el-icon><Operation /></el-icon><span>生成策略</span></template>
          <el-menu-item index="/task-policy">任务策略</el-menu-item>
          <el-menu-item index="/prompt-optimization">提示词优化</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="resources">
          <template #title><el-icon><Files /></el-icon><span>资源与内容安全</span></template>
          <el-menu-item index="/storage">存储与资源</el-menu-item>
        </el-sub-menu>
        <el-sub-menu index="platform">
          <template #title><el-icon><Setting /></el-icon><span>平台与权限</span></template>
          <el-menu-item index="/platform-settings">站点与邮件</el-menu-item>
          <el-menu-item index="/permissions">管理员与权限</el-menu-item>
          <el-menu-item index="/logs">日志管理</el-menu-item>
          <el-menu-item v-if="authStore.adminInfo?.role === 'super_admin'" index="/database">数据管理</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <h4>{{ currentTitle }}</h4>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><UserFilled /></el-icon>
              {{ username }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  DataAnalysis,
  User,
  Picture,
  Setting,
  Coin,
  Connection,
  Operation,
  Files,
  UserFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => route.path)
const openMenus = computed(() => {
  if (['/works', '/tasks', '/announcements', '/reports'].includes(route.path)) return ['creation']
  if (['/users', '/user-defaults'].includes(route.path)) return ['users']
  if (['/credit-packages', '/redeem-codes'].includes(route.path)) return ['credits']
  if (['/accounts'].includes(route.path)) return ['ai']
  if (['/task-policy', '/prompt-optimization'].includes(route.path)) return ['generation']
  if (['/storage'].includes(route.path)) return ['resources']
  if (['/platform-settings', '/permissions', '/logs', '/database'].includes(route.path)) return ['platform']
  return []
})
const currentTitle = computed(() => route.meta.title || '')
const username = computed(() => authStore.adminInfo?.name || authStore.adminInfo?.username || 'admin')

const handleCommand = (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      authStore.logout()
      router.push('/login')
      ElMessage.success('已退出登录')
    }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/profile')
  }
}
</script>

<style scoped>
.layout {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
}

.sidebar :deep(.el-menu) {
  border-right: 0;
}

.sidebar :deep(.el-sub-menu__title),
.sidebar :deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
}

.sidebar :deep(.el-sub-menu .el-menu-item) {
  min-width: auto;
  padding-left: 52px !important;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60px;
  background-color: #2a3f54;
}

.logo h3 {
  color: #fff;
  margin: 0;
  font-size: 18px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 20px;
}

.header-left h4 {
  margin: 0;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
}

.user-info .el-icon {
  margin-right: 5px;
}

.main {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>
