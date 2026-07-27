<template>
  <div class="user-management">
    <div class="header">
      <div>
        <h2>用户管理</h2>
        <p>查看用户资料、会员等级与灵感值余额。</p>
      </div>
      <div class="header-actions">
        <el-button :loading="loading" @click="fetchUsers">刷新</el-button>
        <el-button type="primary" @click="handleAdd">添加用户</el-button>
      </div>
    </div>

    <el-form :inline="true" class="search-form">
      <el-form-item label="用户名">
        <el-input v-model.trim="searchForm.username" placeholder="请输入用户名" clearable @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model.trim="searchForm.email" placeholder="请输入邮箱" clearable @keyup.enter="handleSearch" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="users" style="width: 100%" v-loading="loading" empty-text="暂无用户数据">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="用户名" width="150" />
      <el-table-column prop="email" label="邮箱" width="200" />
      <el-table-column prop="level" label="会员等级" width="120" />
      <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.is_active ? 'success' : 'info'">{{ row.is_active ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column prop="balance" label="灵感值" width="110" />
      <el-table-column label="注册时间" width="180">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="320">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="primary" plain @click="openRecharge(row)">调整灵感值</el-button>
          <el-button size="small" link type="primary" @click="openCreditLedgers(row)">流水</el-button>
          <el-button size="small" link :type="row.is_active ? 'danger' : 'success'" @click="toggleUserStatus(row)">{{ row.is_active ? '停用' : '启用' }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handlePageChange"
    />

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑用户' : '添加用户'" width="500px" @closed="resetFormValidation">
      <el-form ref="formRef" :model="userForm" :rules="rules" label-width="80px" @submit.prevent>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" :disabled="Boolean(editingId)" :placeholder="editingId ? '' : '请输入邮箱'" />
        </el-form-item>
        <el-form-item label="会员等级" prop="level" v-if="editingId">
          <el-select v-model="userForm.level" style="width: 100%"><el-option label="免费用户" value="free"/><el-option label="会员" value="member"/><el-option label="专业会员" value="pro"/></el-select>
        </el-form-item>
        <el-form-item :label="editingId ? '新密码' : '密码'" prop="password">
          <el-input v-model="userForm.password" type="password" :placeholder="editingId ? '留空表示不修改' : '请输入密码'" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rechargeDialog" title="调整灵感值" width="440px" @closed="resetRechargeForm">
      <el-form :model="rechargeForm" label-width="90px">
        <el-form-item label="用户"><span>{{ rechargeUser?.name }}（{{ rechargeUser?.email }}）</span></el-form-item>
        <el-form-item label="调整方式"><el-radio-group v-model="rechargeForm.type"><el-radio value="add">增加</el-radio><el-radio value="deduct">扣减</el-radio></el-radio-group></el-form-item>
        <el-form-item label="调整数量" required><el-input-number v-model="rechargeForm.amount" :min="1" :max="1000000" class="w-full" /></el-form-item>
        <el-form-item label="调整备注" required><el-input v-model="rechargeForm.remark" maxlength="200" show-word-limit type="textarea" placeholder="例如：活动补偿、违规扣减" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="rechargeDialog = false">取消</el-button><el-button :type="rechargeForm.type === 'deduct' ? 'danger' : 'primary'" :loading="recharging" @click="submitRecharge">确认调整</el-button></template>
    </el-dialog>

    <el-dialog v-model="ledgerDialog" :title="ledgerUser ? `${ledgerUser.name} 的灵感值流水` : '灵感值流水'" width="860px" @closed="resetLedgers">
      <el-form :inline="true" class="ledger-filters">
        <el-form-item label="方向"><el-select v-model="ledgerFilters.type" clearable placeholder="全部" class="ledger-select"><el-option label="收入" value="income" /><el-option label="支出" value="expense" /></el-select></el-form-item>
        <el-form-item label="来源"><el-select v-model="ledgerFilters.source_type" clearable placeholder="全部" class="ledger-select"><el-option label="管理员增加" value="admin_adjustment_add" /><el-option label="管理员扣减" value="admin_adjustment_deduct" /><el-option label="生成任务" value="task" /><el-option label="提示词优化" value="prompt_optimization" /></el-select></el-form-item>
        <el-form-item label="时间范围"><el-date-picker v-model="ledgerFilters.dateRange" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" value-format="YYYY-MM-DD HH:mm:ss" /></el-form-item>
        <el-form-item><el-button type="primary" @click="searchCreditLedgers">筛选</el-button><el-button @click="resetCreditLedgerFilters">重置</el-button></el-form-item>
      </el-form>
      <el-table :data="ledgers" v-loading="ledgerLoading" empty-text="暂无灵感值流水">
        <el-table-column prop="id" label="流水 ID" width="90" />
        <el-table-column label="方向" width="90"><template #default="{ row }"><el-tag :type="row.type === 'income' ? 'success' : 'danger'">{{ row.type === 'income' ? '收入' : '支出' }}</el-tag></template></el-table-column>
        <el-table-column label="数量" width="90"><template #default="{ row }"><span :class="row.type === 'income' ? 'income' : 'expense'">{{ row.type === 'income' ? '+' : '-' }}{{ row.amount }}</span></template></el-table-column>
        <el-table-column prop="balance_after" label="调整后余额" width="110" />
        <el-table-column prop="title" label="说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="来源" min-width="150"><template #default="{ row }">{{ sourceText(row.source_type) }}<span v-if="row.source_id"> #{{ row.source_id }}</span></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
      <el-pagination v-model:current-page="ledgerPagination.page" v-model:page-size="ledgerPagination.pageSize" :total="ledgerPagination.total" layout="total, prev, pager, next" @current-change="fetchCreditLedgers" />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const users = ref([])
const searchForm = ref({
  username: '',
  email: ''
})
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const editingId = ref(0)
const submitting = ref(false)
const formRef = ref(null)
const rechargeDialog = ref(false)
const recharging = ref(false)
const rechargeUser = ref(null)
const rechargeForm = reactive({ type: 'add', amount: 100, remark: '' })
const ledgerDialog = ref(false)
const ledgerLoading = ref(false)
const ledgerUser = ref(null)
const ledgers = ref([])
const ledgerPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const ledgerFilters = reactive({ type: '', source_type: '', dateRange: [] })
const userForm = reactive({
  username: '',
  email: '',
  password: '',
  level: 'free'
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { validator: (_, value, callback) => {
      if (!editingId.value && !value) return callback(new Error('请输入密码'))
      if (value && value.length < 8) return callback(new Error('密码至少8位'))
      callback()
    }, trigger: 'blur' }
  ]
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const data = await request.get('/users', {
      params: {
        page: pagination.value.page,
        page_size: pagination.value.pageSize,
        ...searchForm.value
      }
    })
    users.value = data?.list || []
    pagination.value.total = data?.total || 0
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  fetchUsers()
}

const handleReset = () => {
  searchForm.value = {
    username: '',
    email: ''
  }
  handleSearch()
}

const handleAdd = () => {
  editingId.value = 0
  dialogVisible.value = true
  Object.assign(userForm, {
    username: '',
    email: '',
    password: '',
    level: 'free'
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return

  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingId.value) {
      const payload = {
        username: userForm.username,
        level: userForm.level
      }
      if (userForm.password) payload.password = userForm.password
      await request.put(`/users/${editingId.value}`, payload)
      ElMessage.success('用户信息已更新')
    } else {
      await request.post('/users', {
        username: userForm.username,
        email: userForm.email,
        password: userForm.password
      })
      ElMessage.success('创建用户成功')
    }
    dialogVisible.value = false
    await fetchUsers()
  } catch {
    // 错误已在响应拦截器中提示。
  } finally {
    submitting.value = false
  }
}

const handleEdit = (row) => {
  editingId.value = row.id
  Object.assign(userForm, { username: row.name, email: row.email, password: '', level: row.level || 'free' })
  dialogVisible.value = true
}

const toggleUserStatus = async (row) => {
  const isActive = !row.is_active
  const action = isActive ? '启用' : '停用'
  const detail = isActive ? '用户将可以重新登录。' : '该用户的全部登录会话将立即失效。'
  try {
    await ElMessageBox.confirm(`确定${action}用户“${row.name}”吗？${detail}`, '确认操作', { type: 'warning' })
    await request.put(`/users/${row.id}/status`, { is_active: isActive })
    ElMessage.success(`用户已${action}`)
    await fetchUsers()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      // 错误已在响应拦截器中提示。
    }
  }
}

const openRecharge = (row) => {
  rechargeUser.value = row
  Object.assign(rechargeForm, { type: 'add', amount: 100, remark: '' })
  rechargeDialog.value = true
}

const sourceText = (source) => ({
  admin_adjustment_add: '管理员增加',
  admin_adjustment_deduct: '管理员扣减',
  task: '生成任务',
  prompt_optimization: '提示词优化',
  register: '注册赠送',
  gift: '赠送',
  recharge: '充值',
  checkin: '签到',
  redeem: '兑换'
}[source] || source || '-')

const openCreditLedgers = (row) => {
  ledgerUser.value = row
  ledgerPagination.page = 1
  ledgerPagination.total = 0
  Object.assign(ledgerFilters, { type: '', source_type: '', dateRange: [] })
  ledgers.value = []
  ledgerDialog.value = true
  fetchCreditLedgers()
}

const fetchCreditLedgers = async () => {
  if (!ledgerUser.value) return
  ledgerLoading.value = true
  try {
    const data = await request.get(`/users/${ledgerUser.value.id}/credits`, { params: { page: ledgerPagination.page, page_size: ledgerPagination.pageSize, type: ledgerFilters.type, source_type: ledgerFilters.source_type, start_at: ledgerFilters.dateRange?.[0] || '', end_at: ledgerFilters.dateRange?.[1] || '' } })
    ledgers.value = data.list || []
    ledgerPagination.total = data.total || 0
  } finally {
    ledgerLoading.value = false
  }
}

const searchCreditLedgers = () => {
  ledgerPagination.page = 1
  fetchCreditLedgers()
}

const resetCreditLedgerFilters = () => {
  Object.assign(ledgerFilters, { type: '', source_type: '', dateRange: [] })
  searchCreditLedgers()
}

const submitRecharge = async () => {
  if (!rechargeUser.value) return
  if (!rechargeForm.amount || rechargeForm.amount < 1) return ElMessage.warning('请输入有效充值数量')
  if (!rechargeForm.remark.trim()) return ElMessage.warning('请填写充值备注')
  recharging.value = true
  try {
    await request.post(`/users/${rechargeUser.value.id}/credits`, {
      type: rechargeForm.type,
      amount: rechargeForm.amount,
      remark: rechargeForm.remark.trim()
    })
    rechargeDialog.value = false
    ElMessage.success(rechargeForm.type === 'deduct' ? '灵感值已扣减' : '灵感值已增加')
    await fetchUsers()
  } finally {
    recharging.value = false
  }
}

const handleSizeChange = () => {
  pagination.value.page = 1
  fetchUsers()
}

const handlePageChange = () => {
  fetchUsers()
}

const resetFormValidation = () => {
  formRef.value?.clearValidate()
}

const resetRechargeForm = () => {
  rechargeUser.value = null
  Object.assign(rechargeForm, { type: 'add', amount: 100, remark: '' })
}

const resetLedgers = () => {
  ledgerUser.value = null
  ledgers.value = []
  ledgerPagination.page = 1
  ledgerPagination.total = 0
  Object.assign(ledgerFilters, { type: '', source_type: '', dateRange: [] })
}

const formatDate = (value) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (number) => String(number).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
}

.header p {
  margin: 6px 0 0;
  color: #909399;
  font-size: 13px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.search-form {
  margin-bottom: 20px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
.income { color: #67c23a; font-weight: 600; }
.expense { color: #f56c6c; font-weight: 600; }
.ledger-filters { margin-bottom: 12px; }
.ledger-select { width: 130px; }
</style>
