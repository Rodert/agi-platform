<template>
  <div class="user-defaults" v-loading="loading">
    <h2>用户默认设置</h2>
    <el-form :model="form" label-width="160px" class="settings-form">
      <el-form-item label="新用户礼包（灵感值）">
        <el-input-number v-model="form.new_user_gift_amount" :min="0" :max="1000000" :precision="0" />
      </el-form-item>
      <el-form-item label="默认用户等级"><el-select v-model="form.default_user_level"><el-option label="免费版" value="free" /><el-option label="会员版" value="member" /><el-option label="专业版" value="pro" /></el-select></el-form-item>
      <el-form-item label="默认头像">
        <el-radio-group v-model="avatarMode"><el-radio value="system">系统默认头像</el-radio><el-radio value="custom">自定义图片</el-radio></el-radio-group>
      </el-form-item>
      <el-form-item v-if="avatarMode === 'custom'" label="图片地址"><el-input v-model="form.default_user_avatar" placeholder="https://example.com/avatar.png" /></el-form-item>
      <el-form-item label="注册邮箱验证"><el-switch v-model="form.register_email_verification" /></el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const saving = ref(false)
const form = reactive({ new_user_gift_amount: 5, default_user_level: 'free', default_user_avatar: '', register_email_verification: true })
const avatarMode = ref('system')

async function load() {
  loading.value = true
  try {
    const data = await request.get('/config/user-defaults')
    Object.assign(form, { new_user_gift_amount: Number(data.new_user_gift_amount ?? 5), default_user_level: data.default_user_level || 'free', default_user_avatar: data.default_user_avatar || '', register_email_verification: data.register_email_verification !== false })
    avatarMode.value = form.default_user_avatar ? 'custom' : 'system'
  } finally {
    loading.value = false
  }
}

watch(avatarMode, mode => { if (mode === 'system') form.default_user_avatar = '' })

async function save() {
  saving.value = true
  try {
    await request.put('/config/user-defaults', form)
    ElMessage.success('用户默认设置已保存')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.user-defaults { padding: 20px; }
.settings-form { max-width: 600px; }
h2 { margin-bottom: 20px; }
</style>
