<template>
  <div class="logs-management">
    <h2>日志管理</h2>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="操作日志" name="operation">
        <el-form :inline="true" class="search-form">
          <el-form-item label="操作人">
            <el-input v-model="searchForm.operator" placeholder="请输入用户名" clearable />
          </el-form-item>
          <el-form-item label="操作类型">
            <el-select v-model="searchForm.type" placeholder="请选择" clearable>
              <el-option label="全部" value="" />
              <el-option label="登录" value="login" />
              <el-option label="创建" value="create" />
              <el-option label="更新" value="update" />
              <el-option label="删除" value="delete" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="searchForm.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>

        <el-table :data="operationLogs" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="operator" label="操作人" width="120" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="getTypeColor(row.type)">{{ getTypeName(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="module" label="模块" width="120" />
          <el-table-column prop="action" label="操作" width="150" />
          <el-table-column prop="ip" label="IP地址" width="140" />
          <el-table-column prop="created_at" label="操作时间" width="180" />
          <el-table-column label="操作" fixed="right" width="100">
            <template #default="{ row }">
              <el-button size="small" @click="viewDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, prev, pager, next"
          @current-change="fetchLogs"
        />
      </el-tab-pane>

      <el-tab-pane label="错误日志" name="error">
        <el-table :data="errorLogs" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="level" label="级别" width="100">
            <template #default="{ row }">
              <el-tag :type="getLevelColor(row.level)">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="错误信息" min-width="300" />
          <el-table-column prop="module" label="模块" width="120" />
          <el-table-column prop="created_at" label="时间" width="180" />
          <el-table-column label="操作" fixed="right" width="100">
            <template #default="{ row }">
              <el-button size="small" @click="viewError(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="登录日志" name="login">
        <el-table :data="loginLogs" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="ip" label="IP地址" width="140" />
          <el-table-column prop="location" label="登录地点" width="150" />
          <el-table-column prop="device" label="设备信息" min-width="200" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.success ? 'success' : 'danger'">
                {{ row.success ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="180" />
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('operation')
const loading = ref(false)
const searchForm = ref({
  operator: '',
  type: '',
  dateRange: []
})

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

const operationLogs = ref([
  {
    id: 1,
    operator: 'admin',
    type: 'login',
    module: '系统',
    action: '管理员登录',
    ip: '192.168.1.100',
    created_at: '2026-07-12 20:00:00'
  },
  {
    id: 2,
    operator: 'admin',
    type: 'update',
    module: '用户管理',
    action: '更新用户状态',
    ip: '192.168.1.100',
    created_at: '2026-07-12 20:05:00'
  }
])

const errorLogs = ref([
  {
    id: 1,
    level: 'ERROR',
    message: '数据库连接超时',
    module: 'database',
    created_at: '2026-07-12 19:30:00'
  }
])

const loginLogs = ref([
  {
    id: 1,
    username: 'admin',
    ip: '192.168.1.100',
    location: '北京',
    device: 'Chrome 120 / MacOS',
    success: true,
    created_at: '2026-07-12 20:00:00'
  }
])

const getTypeColor = (type) => {
  const map = {
    login: 'success',
    create: 'primary',
    update: 'warning',
    delete: 'danger'
  }
  return map[type] || 'info'
}

const getTypeName = (type) => {
  const map = {
    login: '登录',
    create: '创建',
    update: '更新',
    delete: '删除'
  }
  return map[type] || type
}

const getLevelColor = (level) => {
  const map = {
    ERROR: 'danger',
    WARN: 'warning',
    INFO: 'info'
  }
  return map[level] || 'info'
}

const fetchLogs = () => {
  loading.value = true
  setTimeout(() => {
    loading.value = false
  }, 500)
}

const handleSearch = () => {
  pagination.value.page = 1
  fetchLogs()
}

const handleReset = () => {
  searchForm.value = {
    operator: '',
    type: '',
    dateRange: []
  }
  handleSearch()
}

const viewDetail = (row) => {
  ElMessage.info(`查看日志详情: ${row.id}`)
}

const viewError = (row) => {
  ElMessage.info(`查看错误详情: ${row.id}`)
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.logs-management {
  padding: 20px;
}

.search-form {
  margin-bottom: 20px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
