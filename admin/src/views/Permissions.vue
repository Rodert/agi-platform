<template>
  <div class="permissions">
    <div class="page-head"><div><h2>权限管理</h2><p>管理后台管理员及其职责范围</p></div><el-button v-if="isSuperAdmin" type="primary" @click="openCreate">添加管理员</el-button></div>

    <el-alert v-if="!isSuperAdmin" title="当前账号仅可查看角色权限说明，管理员账号管理仅限超级管理员。" type="info" :closable="false" class="notice" />
    <el-tabs v-model="activeTab">
      <el-tab-pane label="角色权限" name="roles">
        <el-table :data="roleDefinitions" style="width:100%">
          <el-table-column prop="name" label="角色" width="150"><template #default="{ row }"><el-tag :type="row.role === 'super_admin' ? 'danger' : row.role === 'admin' ? 'primary' : 'success'">{{ row.name }}</el-tag></template></el-table-column>
          <el-table-column prop="description" label="职责" min-width="220" />
          <el-table-column label="可操作模块" min-width="360"><template #default="{ row }"><el-tag v-for="item in row.permissions" :key="item" class="permission-tag">{{ item }}</el-tag></template></el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="管理员账号" name="admins">
        <el-table :data="admins" v-loading="loading" style="width:100%">
          <el-table-column prop="username" label="用户名" min-width="140" />
          <el-table-column prop="name" label="姓名" min-width="130" />
          <el-table-column label="角色" width="140"><template #default="{ row }"><el-tag :type="roleType(row.role)">{{ roleName(row.role) }}</el-tag></template></el-table-column>
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.is_active ? 'success' : 'info'">{{ row.is_active ? '正常' : '已禁用' }}</el-tag></template></el-table-column>
          <el-table-column prop="last_login_at" label="最后登录" min-width="170"><template #default="{ row }">{{ row.last_login_at || '从未登录' }}</template></el-table-column>
          <el-table-column label="操作" width="180" fixed="right"><template #default="{ row }"><el-button v-if="isSuperAdmin && row.role !== 'super_admin'" size="small" @click="openEdit(row)">编辑</el-button><el-button v-if="isSuperAdmin && row.role !== 'super_admin'" size="small" :type="row.is_active ? 'danger' : 'success'" @click="toggleActive(row)">{{ row.is_active ? '禁用' : '启用' }}</el-button><span v-if="row.role === 'super_admin'" class="muted">受保护账号</span></template></el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="dialogOpen" :title="editing ? '编辑管理员' : '添加管理员'" width="460px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="用户名" prop="username"><el-input v-model="form.username" :disabled="editing" autocomplete="off" /></el-form-item>
        <el-form-item label="姓名" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="角色" prop="role"><el-select v-model="form.role" class="full"><el-option label="管理员" value="admin" /><el-option label="审核员" value="auditor" /></el-select></el-form-item>
        <el-form-item :label="editing ? '重置密码' : '登录密码'" prop="password"><el-input v-model="form.password" type="password" show-password :placeholder="editing ? '留空则不修改' : '至少 8 位'" autocomplete="new-password" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogOpen=false">取消</el-button><el-button type="primary" :loading="saving" @click="submit">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createAdmin, getAdmins, updateAdmin } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const activeTab = ref('roles'), admins = ref([]), loading = ref(false), saving = ref(false), dialogOpen = ref(false), editing = ref(null), formRef = ref()
const isSuperAdmin = computed(() => authStore.adminInfo?.role === 'super_admin')
const form = reactive({ username: '', name: '', role: 'auditor', password: '' })
const rules = { username: [{ required: true, message: '请输入用户名', trigger: 'blur' }], name: [{ required: true, message: '请输入姓名', trigger: 'blur' }], role: [{ required: true, message: '请选择角色', trigger: 'change' }] }
const roleDefinitions = [
  { role: 'super_admin', name: '超级管理员', description: '拥有平台全部管理权限，可管理管理员账号。', permissions: ['平台配置', '渠道账号', '用户与灵感值', '作品审核', '公告报表', '管理员管理'] },
  { role: 'admin', name: '管理员', description: '负责日常运营和业务配置，不可管理管理员账号。', permissions: ['平台配置', '渠道账号', '用户与灵感值', '作品审核', '公告报表'] },
  { role: 'auditor', name: '审核员', description: '负责内容审核和运营数据查看。', permissions: ['作品审核', '生成记录', '公告查看', '数据报表'] },
]
const roleName = role => ({ super_admin: '超级管理员', admin: '管理员', auditor: '审核员' }[role] || role)
const roleType = role => ({ super_admin: 'danger', admin: 'primary', auditor: 'success' }[role] || 'info')
const resetForm = () => { editing.value = null; Object.assign(form, { username: '', name: '', role: 'auditor', password: '' }); formRef.value?.clearValidate() }
const load = async () => { loading.value = true; try { admins.value = await getAdmins() } finally { loading.value = false } }
const openCreate = () => { dialogOpen.value = true }
const openEdit = row => { editing.value = row; Object.assign(form, { username: row.username, name: row.name, role: row.role, password: '' }); dialogOpen.value = true }
const submit = async () => { await formRef.value.validate(); saving.value = true; try { const payload = { name: form.name, role: form.role, ...(form.password ? { password: form.password } : {}) }; if (editing.value) await updateAdmin(editing.value.id, payload); else { if (!form.password || form.password.length < 8) { ElMessage.warning('密码至少 8 位'); return }; await createAdmin({ username: form.username, ...payload }) }; ElMessage.success('已保存'); dialogOpen.value = false; await load() } finally { saving.value = false } }
const toggleActive = async row => { const next = !row.is_active; await ElMessageBox.confirm(`确定${next ? '启用' : '禁用'}管理员“${row.name}”吗？`, '确认操作', { type: 'warning' }); await updateAdmin(row.id, { name: row.name, role: row.role, is_active: next }); ElMessage.success(next ? '已启用' : '已禁用'); await load() }
onMounted(load)
</script>

<style scoped>
.permissions { padding: 4px; }.page-head { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 20px; }.page-head h2 { margin: 0; }.page-head p,.muted { color: #909399; font-size: 13px; }.page-head p { margin: 6px 0 0; }.notice { margin-bottom: 16px; }.permission-tag { margin: 3px; }.full { width: 100%; }
</style>
