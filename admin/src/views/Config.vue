<template>
  <div class="config-management" v-loading="loading">
    <h2>系统配置</h2>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="基础配置" name="basic">
        <el-form :model="basicConfig" label-width="130px" class="config-form">
          <el-form-item label="网站名称"><el-input v-model="basicConfig.site_name" /></el-form-item>
          <el-form-item label="网站描述"><el-input v-model="basicConfig.site_desc" type="textarea" :rows="3" /></el-form-item>
          <el-form-item label="注册开关"><el-switch v-model="basicConfig.register_enabled" /></el-form-item>
          <el-form-item label="新用户赠送积分"><el-input-number v-model="basicConfig.register_credits" :min="0" /></el-form-item>
          <el-form-item><el-button type="primary" :loading="saving" @click="saveBasicConfig">保存</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="邮件配置" name="email">
        <el-form :model="emailConfig" label-width="130px" class="config-form">
          <el-form-item label="SMTP服务器"><el-input v-model="emailConfig.smtp_host" /></el-form-item>
          <el-form-item label="SMTP端口"><el-input-number v-model="emailConfig.smtp_port" :min="1" :max="65535" /></el-form-item>
          <el-form-item label="SMTP账号"><el-input v-model="emailConfig.smtp_user" /></el-form-item>
          <el-form-item label="SMTP密码"><el-input v-model="emailConfig.smtp_password" type="password" show-password placeholder="留空表示不修改" /></el-form-item>
          <el-form-item label="发件人名称"><el-input v-model="emailConfig.from_name" /></el-form-item>
          <el-form-item label="发件人邮箱"><el-input v-model="emailConfig.from_email" /></el-form-item>
          <el-form-item label="启用SSL"><el-switch v-model="emailConfig.smtp_ssl" /></el-form-item>
          <el-form-item label="启用邮件服务"><el-switch v-model="emailConfig.is_active" /></el-form-item>
          <el-form-item><el-button type="primary" :loading="saving" @click="saveEmailConfig">保存</el-button><el-button @click="testEmail">测试发送</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="存储配置" name="storage">
        <el-button type="primary" class="mb" @click="addStorage">添加存储配置</el-button>
        <el-table :data="storageList">
          <el-table-column prop="name" label="名称" /><el-table-column label="类型"><template #default="{ row }"><el-tag>{{ getStorageTypeName(row.type) }}</el-tag></template></el-table-column>
          <el-table-column prop="domain" label="访问域名" /><el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.is_enabled ? 'success' : 'info'">{{ row.is_enabled ? '已启用' : '未启用' }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="220"><template #default="{ row }"><el-button size="small" @click="editStorage(row)">编辑</el-button><el-button v-if="!row.is_enabled" size="small" type="primary" @click="enableStorage(row)">启用</el-button><el-button v-if="!row.is_enabled" size="small" type="danger" @click="deleteStorage(row)">删除</el-button></template></el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="storageDialog" :title="storageForm.id ? '编辑存储配置' : '添加存储配置'" width="600px"><el-form :model="storageForm" label-width="110px"><el-form-item label="名称"><el-input v-model="storageForm.name" /></el-form-item><el-form-item label="类型"><el-select v-model="storageForm.type"><el-option label="本地存储" value="local"/><el-option label="腾讯云 COS" value="tencent_cos"/><el-option label="阿里云 OSS" value="aliyun_oss"/><el-option label="Cloudflare R2" value="cloudflare"/></el-select></el-form-item><el-form-item v-if="storageForm.type==='local'" label="本地路径"><el-input v-model="storageForm.local_path" /></el-form-item><template v-else><el-form-item label="Endpoint"><el-input v-model="storageForm.endpoint" /></el-form-item><el-form-item label="Access Key"><el-input v-model="storageForm.access_key" /></el-form-item><el-form-item label="Secret Key"><el-input v-model="storageForm.secret_key" type="password" placeholder="编辑时留空表示不修改" /></el-form-item><el-form-item label="Bucket"><el-input v-model="storageForm.bucket" /></el-form-item><el-form-item label="Region"><el-input v-model="storageForm.region" /></el-form-item></template><el-form-item label="访问域名"><el-input v-model="storageForm.domain" /></el-form-item></el-form><template #footer><el-button @click="storageDialog=false">取消</el-button><el-button type="primary" @click="saveStorage">保存</el-button></template></el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const activeTab=ref('basic'),loading=ref(false),saving=ref(false),storageList=ref([]),storageDialog=ref(false)
const basicConfig=reactive({site_name:'',site_desc:'',register_enabled:true,register_credits:0})
const emailConfig=reactive({smtp_host:'',smtp_port:587,smtp_user:'',smtp_password:'',smtp_ssl:false,from_name:'',from_email:'',is_active:false})
const emptyStorage=()=>({id:0,name:'',type:'local',local_path:'./uploads',endpoint:'',access_key:'',secret_key:'',bucket:'',region:'',domain:''})
const storageForm=reactive(emptyStorage())
const bool=v=>v===true||v==='true'

async function loadAll(){loading.value=true;try{const [basic,email,storage]=await Promise.all([request.get('/config/basic'),request.get('/config/email'),request.get('/config/storage')]);Object.assign(basicConfig,{site_name:basic.site_name||'AGI Platform',site_desc:basic.site_desc||'',register_enabled:bool(basic.register_enabled),register_credits:Number(basic.new_user_gift_amount||0)});Object.assign(emailConfig,email,{smtp_password:''});storageList.value=storage||[]}finally{loading.value=false}}
async function saveBasicConfig(){saving.value=true;try{await request.put('/config/basic',basicConfig);ElMessage.success('基础配置已保存')}finally{saving.value=false}}
async function saveEmailConfig(){saving.value=true;try{await request.put('/config/email',emailConfig);emailConfig.smtp_password='';ElMessage.success('邮件配置已保存')}finally{saving.value=false}}
async function testEmail(){const {value}=await ElMessageBox.prompt('请输入测试邮件接收地址','测试邮件',{inputPattern:/^[^\s@]+@[^\s@]+\.[^\s@]+$/,inputErrorMessage:'请输入有效邮箱'});await request.post('/config/email/test',{email:value});ElMessage.success('测试邮件已发送')}
function addStorage(){Object.assign(storageForm,emptyStorage());storageDialog.value=true}
function editStorage(row){Object.assign(storageForm,emptyStorage(),row,{secret_key:''});storageDialog.value=true}
async function saveStorage(){const body={...storageForm};delete body.id;if(storageForm.id)await request.put(`/config/storage/${storageForm.id}`,body);else await request.post('/config/storage',body);storageDialog.value=false;ElMessage.success('存储配置已保存');await loadAll()}
async function enableStorage(row){await request.post(`/config/storage/${row.id}/enable`);ElMessage.success('存储已启用');await loadAll()}
async function deleteStorage(row){await ElMessageBox.confirm(`确认删除“${row.name}”？`,'删除配置',{type:'warning'});await request.delete(`/config/storage/${row.id}`);ElMessage.success('删除成功');await loadAll()}
const getStorageTypeName=type=>({local:'本地存储',tencent_cos:'腾讯云COS',aliyun_oss:'阿里云OSS',cloudflare:'Cloudflare R2'}[type]||type)
onMounted(loadAll)
</script>

<style scoped>.config-management{padding:20px}.config-form{max-width:680px}.mb{margin-bottom:20px}h2{margin-bottom:20px}</style>
