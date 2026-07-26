<template>
  <div class="logs-management">
    <h2>日志管理</h2>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="操作日志" name="operation">
        <el-form :inline="true" class="search-form">
          <el-form-item label="操作人"><el-input v-model="searchForm.operator" placeholder="用户名或姓名" clearable @keyup.enter="handleSearch" /></el-form-item>
          <el-form-item label="操作类型"><el-select v-model="searchForm.action" placeholder="全部" clearable><el-option label="登录成功" value="login" /><el-option label="登录失败" value="login_failed" /><el-option label="用户创建" value="create_user" /><el-option label="用户更新" value="update_user" /><el-option label="作品审核" value="audit_work" /><el-option label="积分调整" value="adjust_credit" /><el-option label="作品下架" value="offline_work" /><el-option label="重新上架" value="republish_work" /><el-option label="验证码已发送" value="send_verification_code" /><el-option label="验证码发送失败" value="send_verification_code_failed" /><el-option label="验证码校验成功" value="verify_verification_code" /><el-option label="验证码校验失败" value="verify_verification_code_failed" /></el-select></el-form-item>
          <el-form-item label="时间范围"><el-date-picker v-model="searchForm.dateRange" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item>
          <el-form-item><el-button type="primary" @click="handleSearch">搜索</el-button><el-button @click="handleReset">重置</el-button></el-form-item>
        </el-form>
        <el-table :data="operationLogs" v-loading="loading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column label="操作人" width="140"><template #default="{ row }">{{ operatorName(row) }}</template></el-table-column>
          <el-table-column label="类型" width="120"><template #default="{ row }"><el-tag :type="getTypeColor(row.action)">{{ getTypeName(row.action) }}</el-tag></template></el-table-column>
          <el-table-column label="对象" width="140"><template #default="{ row }">{{ targetText(row) }}</template></el-table-column>
          <el-table-column prop="description" label="操作说明" min-width="180" show-overflow-tooltip />
          <el-table-column prop="ip" label="IP 地址" width="140" />
          <el-table-column prop="created_at" label="操作时间" width="180" />
          <el-table-column label="操作" fixed="right" width="80"><template #default="{ row }"><el-button link type="primary" @click="viewDetail(row)">详情</el-button></template></el-table-column>
        </el-table>
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total, prev, pager, next" @current-change="fetchLogs" />
      </el-tab-pane>

      <el-tab-pane label="错误日志" name="error"><el-empty description="系统错误日志尚未配置采集与存储，不展示模拟数据。" /></el-tab-pane>

      <el-tab-pane label="登录日志" name="login">
        <el-form :inline="true" class="search-form"><el-form-item label="账号"><el-input v-model="loginSearch.operator" placeholder="用户名或姓名" clearable @keyup.enter="searchLogin" /></el-form-item><el-form-item label="状态"><el-select v-model="loginSearch.action" class="login-status"><el-option label="全部" value="" /><el-option label="成功" value="login" /><el-option label="失败" value="login_failed" /></el-select></el-form-item><el-form-item label="时间范围"><el-date-picker v-model="loginSearch.dateRange" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item><el-form-item><el-button type="primary" @click="searchLogin">搜索</el-button><el-button @click="resetLogin">重置</el-button></el-form-item></el-form>
        <el-table :data="loginLogs" v-loading="loading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column label="用户名" width="150"><template #default="{ row }">{{ operatorName(row) }}</template></el-table-column>
          <el-table-column prop="ip" label="IP 地址" width="140" />
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.action === 'login' ? 'success' : 'danger'">{{ row.action === 'login' ? '成功' : '失败' }}</el-tag></template></el-table-column>
          <el-table-column prop="created_at" label="时间" width="180" />
          <el-table-column label="操作" fixed="right" width="80"><template #default="{ row }"><el-button link type="primary" @click="viewDetail(row)">详情</el-button></template></el-table-column>
        </el-table>
        <el-pagination v-model:current-page="loginPagination.page" v-model:page-size="loginPagination.pageSize" :total="loginPagination.total" layout="total, prev, pager, next" @current-change="fetchLogs" />
      </el-tab-pane>
    </el-tabs>
    <el-dialog v-model="detailVisible" title="操作日志详情" width="680px">
      <el-descriptions v-if="selected" :column="2" border>
        <el-descriptions-item label="日志 ID">{{ selected.id }}</el-descriptions-item><el-descriptions-item label="操作人">{{ operatorName(selected) }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">{{ getTypeName(selected.action) }}</el-descriptions-item><el-descriptions-item label="对象">{{ targetText(selected) }}</el-descriptions-item>
        <el-descriptions-item label="IP 地址">{{ selected.ip || '-' }}</el-descriptions-item><el-descriptions-item label="操作时间">{{ selected.created_at }}</el-descriptions-item>
        <el-descriptions-item label="说明" :span="2">{{ selected.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="变更前" :span="2"><pre>{{ formatData(selected.before_data) }}</pre></el-descriptions-item>
        <el-descriptions-item label="变更后" :span="2"><pre>{{ formatData(selected.after_data) }}</pre></el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { getLogs } from '@/api/admin'

const activeTab = ref('operation')
const loading = ref(false)
const searchForm = ref({ operator: '', action: '', dateRange: [] })
const pagination = ref({ page: 1, pageSize: 20, total: 0 })
const loginSearch = ref({ operator: '', action: '', dateRange: [] })
const loginPagination = ref({ page: 1, pageSize: 20, total: 0 })
const operationLogs = ref([])
const loginLogs = ref([])
const detailVisible = ref(false)
const selected = ref(null)

const getTypeColor = type => ({ login: 'success', login_failed: 'danger', create_user: 'primary', update_user: 'warning', audit_work: 'warning', adjust_credit: 'primary', offline_work: 'danger', republish_work: 'success', send_verification_code: 'success', send_verification_code_failed: 'danger', verify_verification_code: 'success', verify_verification_code_failed: 'danger' }[type] || 'info')
const getTypeName = type => ({ login: '登录成功', login_failed: '登录失败', create_user: '创建用户', update_user: '更新用户', audit_work: '作品审核', adjust_credit: '积分调整', offline_work: '作品下架', republish_work: '重新上架', send_verification_code: '验证码已发送', send_verification_code_failed: '验证码发送失败', verify_verification_code: '验证码校验成功', verify_verification_code_failed: '验证码校验失败' }[type] || type)
const operatorName = row => row.target_type === 'email_verification' ? '用户认证' : (row.admin?.name || row.admin?.username || `管理员 #${row.admin_id}`)
const targetText = row => row.target_type === 'email_verification' ? '邮箱验证码' : (row.target_type ? `${row.target_type}${row.target_id ? ` #${row.target_id}` : ''}` : '系统')
const formatData = value => { try { return JSON.stringify(typeof value === 'string' ? JSON.parse(value || '{}') : value || {}, null, 2) } catch { return value || '{}' } }

async function fetchLogs() {
  loading.value = true
  try {
    const isLogin = activeTab.value === 'login'
    const filters = isLogin ? loginSearch.value : searchForm.value
    const pager = isLogin ? loginPagination.value : pagination.value
    const params = { page: pager.page, page_size: pager.pageSize, login_only: isLogin }
    if (isLogin && filters.action) params.action = filters.action
    else {
      params.operator = filters.operator
      params.action = filters.action
      params.start_at = filters.dateRange?.[0] || ''
      params.end_at = filters.dateRange?.[1] || ''
    }
    if (isLogin) { params.operator = filters.operator; params.start_at = filters.dateRange?.[0] || ''; params.end_at = filters.dateRange?.[1] || '' }
    const data = await getLogs(params)
    if (isLogin) loginLogs.value = data.list || []
    else operationLogs.value = data.list || []
    pager.total = data.total || 0
  } finally { loading.value = false }
}

function handleSearch() { pagination.value.page = 1; fetchLogs() }
function handleReset() { searchForm.value = { operator: '', action: '', dateRange: [] }; handleSearch() }
function searchLogin() { loginPagination.value.page = 1; fetchLogs() }
function resetLogin() { loginSearch.value = { operator: '', action: '', dateRange: [] }; searchLogin() }
function viewDetail(row) { selected.value = row; detailVisible.value = true }

watch(activeTab, tab => { if (tab !== 'error') { (tab === 'login' ? loginPagination : pagination).value.page = 1; fetchLogs() } })
onMounted(fetchLogs)
</script>

<style scoped>
.logs-management { padding: 20px; }
.search-form { margin-bottom: 20px; }
.login-status { width: 110px; }
.el-pagination { margin-top: 20px; justify-content: flex-end; }
pre { margin: 0; white-space: pre-wrap; font-size: 12px; }
</style>
