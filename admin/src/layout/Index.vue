<template>
  <el-container class="layout">
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <div>
          <h3>AGI Platform</h3>
          <el-button class="version-check" link @click="versionDialogVisible = true; refreshVersions()">
            <el-icon><RefreshRight /></el-icon>
            v{{ APP_VERSION }}
            <el-badge v-if="hasUpdate" is-dot class="update-dot" />
          </el-button>
        </div>
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

  <el-dialog v-model="versionDialogVisible" title="版本管理" width="420px" destroy-on-close>
    <div class="version-dialog-heading">
      <span>当前版本</span>
      <el-button text :loading="checkingUpdate" @click="refreshVersions(true)"><el-icon><RefreshRight /></el-icon>刷新版本</el-button>
    </div>
    <div class="current-version">v{{ APP_VERSION }}</div>
    <div class="version-status">
      <el-icon :class="hasUpdate ? 'has-update' : 'up-to-date'"><CircleCheckFilled /></el-icon>
      {{ hasUpdate ? `发现新版本 v${latestVersion}` : '已是最新版本' }}
    </div>
    <el-button v-if="hasUpdate && updateEnabled" class="deploy-latest" type="primary" @click="startUpdate(latestVersion)">更新并重启至 v{{ latestVersion }}</el-button>

    <el-divider>版本回退</el-divider>
    <p class="rollback-hint">选择要回退到的版本，最多显示最近 3 个历史版本。</p>
    <el-radio-group v-model="selectedRollbackVersion" class="rollback-list">
      <el-radio v-for="release in rollbackVersions" :key="release.version" :value="release.version" border>
        <span>v{{ release.version }}</span>
        <small>{{ release.publishedAt }}</small>
      </el-radio>
    </el-radio-group>
    <el-empty v-if="!checkingUpdate && !rollbackVersions.length" description="暂无可回退版本" :image-size="60" />
    <template #footer>
      <el-button @click="versionDialogVisible = false">关闭</el-button>
      <el-button type="warning" :disabled="!updateEnabled || !selectedRollbackVersion" @click="startUpdate(selectedRollbackVersion)">回滚并重启</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
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
  UserFilled,
  RefreshRight,
  CircleCheckFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { APP_VERSION, RELEASES_URL, compareVersions } from '@/config/release'
import { getSystemUpdateStatus, triggerSystemUpdate } from '@/api/admin'

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
const checkingUpdate = ref(false)
const hasUpdate = ref(false)
const latestVersion = ref('')
const updateEnabled = ref(false)
const versionDialogVisible = ref(false)
const releases = ref([])
const selectedRollbackVersion = ref('')
const rollbackVersions = computed(() => releases.value.filter((release) => release.version !== APP_VERSION).slice(0, 3))

const refreshVersions = async (notifyOnError = false) => {
  if (checkingUpdate.value) return

  checkingUpdate.value = true
  try {
    const response = await fetch(RELEASES_URL, {
      headers: { Accept: 'application/vnd.github+json' }
    })

    if (!response.ok) {
      throw new Error('暂未发布可检测的版本')
    }

    const items = await response.json()
    releases.value = items
      .filter((release) => !release.draft && !release.prerelease && /^v?\d+\.\d+\.\d+$/.test(release.tag_name || ''))
      .map((release) => ({
        version: String(release.tag_name).replace(/^v/, ''),
        publishedAt: new Date(release.published_at).toLocaleDateString('zh-CN')
      }))
      .sort((left, right) => compareVersions(right.version, left.version))
    latestVersion.value = releases.value[0]?.version || APP_VERSION
    hasUpdate.value = Boolean(latestVersion.value && compareVersions(latestVersion.value, APP_VERSION) > 0)
  } catch (error) {
    if (notifyOnError) ElMessage.warning(error.message || '版本检测失败，请稍后重试')
  } finally {
    checkingUpdate.value = false
  }
}

onMounted(async () => {
  refreshVersions()
  if (authStore.adminInfo?.role === 'super_admin') {
    try {
      updateEnabled.value = Boolean((await getSystemUpdateStatus()).enabled)
    } catch {
      updateEnabled.value = false
    }
  }
})

const startUpdate = (version) => {
  if (!version) return
  const action = compareVersions(version, APP_VERSION) >= 0 ? '更新' : '回滚'
  ElMessageBox.confirm(`将${action}并重启至 v${version}，期间后台会短暂不可用。确定继续吗？`, `确认${action}`, {
    confirmButtonText: `开始${action}`,
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await triggerSystemUpdate(version)
    ElMessage.success(`${action}已启动，页面将在几秒后恢复`)
    window.setTimeout(() => window.location.reload(), 10000)
  }).catch(() => {})
}

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

.version-check {
  min-height: 20px;
  padding: 0;
  color: #bfcbd9;
  font-size: 12px;
}

.version-check:hover {
  color: #fff;
}

.version-check .el-icon {
  margin-right: 4px;
}

.update-dot {
  margin-left: 5px;
}

.version-dialog-heading { display: flex; align-items: center; justify-content: space-between; color: #606266; }
.current-version { margin-top: 12px; font-size: 26px; font-weight: 600; color: #303133; }
.version-status { display: flex; align-items: center; gap: 6px; margin-top: 8px; color: #909399; font-size: 13px; }
.up-to-date { color: #67c23a; }
.has-update { color: #e6a23c; }
.deploy-latest { width: 100%; margin-top: 18px; }
.rollback-hint { margin: 0 0 12px; color: #909399; font-size: 13px; }
.rollback-list { display: flex; width: 100%; flex-direction: column; gap: 8px; }
.rollback-list :deep(.el-radio) { display: flex; width: 100%; height: 40px; align-items: center; justify-content: space-between; margin: 0; }
.rollback-list small { margin-left: auto; color: #909399; }

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
