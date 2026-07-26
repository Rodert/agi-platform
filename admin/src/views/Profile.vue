<template>
  <div class="profile-page" v-loading="loading">
    <div class="page-head"><h2>个人中心</h2><p>管理当前管理员资料和登录密码</p></div>
    <el-card class="profile-card" shadow="never">
      <template #header><div class="card-title"><el-icon><UserFilled /></el-icon><span>账号资料</span></div></template>
      <el-descriptions :column="1" border class="details">
        <el-descriptions-item label="登录账号">{{ profile.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="角色"><el-tag>{{ roleName(profile.role) }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="上次登录">{{ profile.last_login_at || '-' }}<span v-if="profile.last_login_ip" class="muted"> · {{ profile.last_login_ip }}</span></el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ profile.created_at || '-' }}</el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="profile-form">
        <el-form-item label="显示名称" prop="name"><el-input v-model="form.name" maxlength="50" show-word-limit /></el-form-item>
        <el-divider content-position="left">修改密码</el-divider>
        <el-alert title="如不修改密码，请保持以下字段为空。修改成功后将退出登录。" type="info" :closable="false" show-icon class="password-tip" />
        <el-form-item label="当前密码" prop="current_password"><el-input v-model="form.current_password" type="password" show-password autocomplete="current-password" /></el-form-item>
        <el-form-item label="新密码" prop="new_password"><el-input v-model="form.new_password" type="password" show-password autocomplete="new-password" /></el-form-item>
        <el-form-item><el-button type="primary" :loading="saving" @click="save">保存修改</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UserFilled } from '@element-plus/icons-vue'
import { getProfile, updateProfile } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const formRef = ref()
const profile = reactive({ username: '', name: '', role: '', last_login_at: '', last_login_ip: '', created_at: '' })
const form = reactive({ name: '', current_password: '', new_password: '' })
const rules = {
  name: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
  new_password: [{ validator: (_, value, callback) => { if ((value || form.current_password) && value.length < 8) callback(new Error('新密码至少 8 位')); else callback() }, trigger: 'blur' }],
  current_password: [{ validator: (_, value, callback) => { if ((value || form.new_password) && !value) callback(new Error('请输入当前密码')); else callback() }, trigger: 'blur' }]
}
const roleName = role => ({ super_admin: '超级管理员', admin: '管理员', auditor: '审核员' }[role] || role || '-')
async function loadProfile() { loading.value = true; try { const data = await getProfile(); Object.assign(profile, data); form.name = data.name } finally { loading.value = false } }
async function save() {
  await formRef.value.validate()
  saving.value = true
  try {
    const passwordChanged = Boolean(form.new_password)
    const data = await updateProfile({ ...form })
    authStore.updateAdminInfo(data)
    Object.assign(profile, data)
    form.current_password = ''
    form.new_password = ''
    ElMessage.success(passwordChanged ? '密码已修改，请重新登录' : '个人资料已保存')
    if (passwordChanged) { authStore.logout(); router.push('/login') }
  } finally { saving.value = false }
}
onMounted(loadProfile)
</script>

<style scoped>
.profile-page{padding:20px}.page-head{margin-bottom:20px}.page-head h2{margin:0}.page-head p,.muted{color:#909399;font-size:13px}.page-head p{margin:8px 0 0}.profile-card{max-width:760px}.card-title{display:flex;align-items:center;gap:8px;font-weight:600}.details{margin-bottom:4px}.profile-form{max-width:560px}.password-tip{margin:0 0 18px}
</style>
