<template>
  <div class="user-management">
    <div class="header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="handleAdd">添加用户</el-button>
    </div>

    <el-form :inline="true" class="search-form">
      <el-form-item label="用户名">
        <el-input v-model="searchForm.username" placeholder="请输入用户名" clearable />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="searchForm.email" placeholder="请输入邮箱" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="users" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="用户名" width="150" />
      <el-table-column prop="email" label="邮箱" width="200" />
      <el-table-column prop="level" label="会员等级" width="120" />
      <el-table-column prop="balance" label="灵感值" width="110" />
      <el-table-column prop="created_at" label="注册时间" width="180" />
      <el-table-column label="操作" fixed="right" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="primary" plain @click="openRecharge(row)">调整灵感值</el-button>
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

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑用户' : '添加用户'" width="500px">
      <el-form :model="userForm" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
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

    <el-dialog v-model="rechargeDialog" title="调整灵感值" width="440px">
      <el-form :model="rechargeForm" label-width="90px">
        <el-form-item label="用户"><span>{{ rechargeUser?.name }}（{{ rechargeUser?.email }}）</span></el-form-item>
        <el-form-item label="调整方式"><el-radio-group v-model="rechargeForm.type"><el-radio value="add">增加</el-radio><el-radio value="deduct">扣减</el-radio></el-radio-group></el-form-item>
        <el-form-item label="调整数量" required><el-input-number v-model="rechargeForm.amount" :min="1" :max="1000000" class="w-full" /></el-form-item>
        <el-form-item label="调整备注" required><el-input v-model="rechargeForm.remark" maxlength="200" show-word-limit type="textarea" placeholder="例如：活动补偿、违规扣减" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="rechargeDialog = false">取消</el-button><el-button :type="rechargeForm.type === 'deduct' ? 'danger' : 'primary'" :loading="recharging" @click="submitRecharge">确认调整</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
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
    users.value = data.list || []
    pagination.value.total = data.total || 0
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

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (editingId.value) {
        await request.put(`/users/${editingId.value}`, userForm)
        ElMessage.success('用户信息已更新')
      } else {
        await request.post('/users', userForm)
        ElMessage.success('创建用户成功')
      }
      dialogVisible.value = false
      fetchUsers()
    } catch (error) {
      // 错误已在拦截器中处理
    } finally {
      submitting.value = false
    }
  })
}

const handleEdit = (row) => {
  editingId.value = row.id
  Object.assign(userForm, { username: row.name, email: row.email, password: '', level: row.level || 'free' })
  dialogVisible.value = true
}

const openRecharge = (row) => {
  rechargeUser.value = row
  Object.assign(rechargeForm, { type: 'add', amount: 100, remark: '' })
  rechargeDialog.value = true
}

const submitRecharge = async () => {
  if (!rechargeUser.value) return
  if (!rechargeForm.amount || rechargeForm.amount < 1) return ElMessage.warning('请输入有效充值数量')
  if (!rechargeForm.remark.trim()) return ElMessage.warning('请填写充值备注')
  recharging.value = true
  try {
    const result = await request.post(`/users/${rechargeUser.value.id}/credits`, rechargeForm)
    rechargeUser.value.balance = result.balance
    rechargeDialog.value = false
    ElMessage.success(rechargeForm.type === 'deduct' ? '灵感值已扣减' : '灵感值已增加')
  } finally {
    recharging.value = false
  }
}

const handleSizeChange = () => {
  fetchUsers()
}

const handlePageChange = () => {
  fetchUsers()
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

.search-form {
  margin-bottom: 20px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
