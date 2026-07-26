<template>
  <div class="config-management" v-loading="loading">
    <h2>{{ pageTitle }}</h2>
    <el-tabs v-model="activeTab">
      <el-tab-pane v-if="hasTab('basic')" label="站点设置" name="basic">
        <el-form :model="basicConfig" label-width="130px" class="config-form">
          <el-form-item label="网站名称"><el-input v-model="basicConfig.site_name" /></el-form-item>
          <el-form-item label="网站描述"><el-input v-model="basicConfig.site_desc" type="textarea" :rows="3" /></el-form-item>
          <el-form-item label="注册开关"><el-switch v-model="basicConfig.register_enabled" /></el-form-item>
          <el-form-item><el-button type="primary" :loading="saving" @click="saveBasicConfig">保存</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane v-if="hasTab('packages')" label="充值套餐" name="packages">
        <p class="package-tip">用户端固定展示三个套餐。购买链接支持支付页、商品页或外部收款页。</p>
        <el-table :data="creditPackages" border>
          <el-table-column label="排序" width="70"><template #default="{ $index }">{{ $index + 1 }}</template></el-table-column>
          <el-table-column label="套餐名称" min-width="120"><template #default="{ row }"><el-input v-model="row.name" maxlength="50" /></template></el-table-column>
          <el-table-column label="价格（元）" width="150"><template #default="{ row }"><el-input-number v-model="row.price" :min="0.01" :precision="2" /></template></el-table-column>
          <el-table-column label="灵感值" width="130"><template #default="{ row }"><el-input-number v-model="row.points" :min="1" :precision="0" /></template></el-table-column>
          <el-table-column label="套餐说明" min-width="170"><template #default="{ row }"><el-input v-model="row.note" maxlength="255" /></template></el-table-column>
          <el-table-column label="购买链接" min-width="230"><template #default="{ row }"><el-input v-model="row.purchase_url" placeholder="https://..." /></template></el-table-column>
          <el-table-column label="推荐" width="75"><template #default="{ row }"><el-switch v-model="row.is_hot" /></template></el-table-column>
          <el-table-column label="启用" width="75"><template #default="{ row }"><el-switch v-model="row.is_active" /></template></el-table-column>
        </el-table>
        <div class="mt"><el-button type="primary" :loading="saving" @click="saveCreditPackages">保存套餐配置</el-button></div>
      </el-tab-pane>

      <el-tab-pane v-if="hasTab('email')" label="邮件服务" name="email">
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

      <el-tab-pane v-if="hasTab('task')" label="任务策略" name="task">
        <el-form :model="taskConfig" label-width="170px" class="config-form">
          <el-form-item label="单用户进行中任务上限"><el-input-number v-model="taskConfig.max_active_tasks" :min="1" :max="1000" /></el-form-item>
          <el-form-item label="提示词上限（字符）"><el-input-number v-model="taskConfig.prompt_max_length" :min="1" :max="50000" /></el-form-item>
          <el-form-item label="失败重试次数"><el-input-number v-model="taskConfig.max_retry_attempts" :min="0" :max="10" /></el-form-item>
          <el-form-item><el-button type="primary" :loading="saving" @click="saveTaskConfig">保存</el-button></el-form-item>
        </el-form>
      </el-tab-pane>

      <el-tab-pane v-if="hasTab('prompt-optimization')" label="提示词优化" name="prompt-optimization">
        <el-form :model="promptOptimizationConfig" label-width="170px" class="config-form">
          <el-form-item label="启用优化"><el-switch v-model="promptOptimizationConfig.is_active" /></el-form-item>
          <el-form-item label="文本模型"><el-select v-model="promptOptimizationConfig.model_name" filterable class="wide" placeholder="先在账号管理中绑定文本模型"><el-option v-for="model in textModels" :key="model.name" :label="`${model.display_name} (${model.name})`" :value="model.name"/></el-select></el-form-item>
          <el-form-item label="最大输入字符"><el-input-number v-model="promptOptimizationConfig.max_input_length" :min="1" :max="50000" /></el-form-item>
          <el-form-item label="每次消耗灵感值"><el-input-number v-model="promptOptimizationConfig.credit_cost" :min="0" :max="100000" /></el-form-item>
          <el-form-item label="每用户每分钟次数"><el-input-number v-model="promptOptimizationConfig.rate_limit_per_minute" :min="1" :max="120" /></el-form-item>
          <el-form-item label="系统提示词"><el-input v-model="promptOptimizationConfig.system_prompt" type="textarea" :rows="8" /></el-form-item>
          <el-form-item><el-button type="primary" :loading="saving" @click="savePromptOptimizationConfig">保存</el-button></el-form-item>
        </el-form>
        <el-divider>优化记录</el-divider>
        <el-table :data="promptOptimizationLogs" max-height="380">
          <el-table-column label="用户" min-width="130"><template #default="{ row }">{{ row.user?.name || '-' }}<div class="subtle">{{ row.user?.email }}</div></template></el-table-column>
          <el-table-column prop="target_type" label="目标" width="90"><template #default="{ row }">{{ row.target_type === 'video' ? '视频' : '图片' }}</template></el-table-column>
          <el-table-column prop="model_name" label="文本模型" min-width="130"/>
          <el-table-column label="渠道" min-width="110"><template #default="{ row }">{{ row.channel?.name || '-' }}</template></el-table-column>
          <el-table-column prop="original_prompt" label="原提示词" min-width="200" show-overflow-tooltip/>
          <el-table-column prop="optimized_prompt" label="优化结果" min-width="240" show-overflow-tooltip/>
          <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">{{ ({ success: '成功', failed: '失败', processing: '处理中' })[row.status] || row.status }}</el-tag></template></el-table-column>
          <el-table-column label="耗时" width="80"><template #default="{ row }">{{ row.latency_ms }}ms</template></el-table-column>
          <el-table-column label="消耗" width="70"><template #default="{ row }">{{ row.credit_cost }}</template></el-table-column>
          <el-table-column label="时间" width="170"><template #default="{ row }">{{ new Date(row.created_at).toLocaleString('zh-CN') }}</template></el-table-column>
        </el-table>
        <el-pagination v-if="promptOptimizationLogTotal" class="mt" background layout="prev, pager, next" :current-page="promptOptimizationLogPage" :page-size="20" :total="promptOptimizationLogTotal" @current-change="loadPromptOptimizationLogs"/>
      </el-tab-pane>

      <el-tab-pane v-if="hasTab('storage')" label="存储配置" name="storage">
        <el-button type="primary" class="mb" @click="addStorage">添加存储配置</el-button>
        <el-table :data="storageList">
          <el-table-column prop="name" label="名称" /><el-table-column label="类型"><template #default="{ row }"><el-tag>{{ getStorageTypeName(row.type) }}</el-tag></template></el-table-column>
          <el-table-column prop="domain" label="访问域名" /><el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.is_enabled ? 'success' : 'info'">{{ row.is_enabled ? '已启用' : '未启用' }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="220"><template #default="{ row }"><el-button size="small" @click="editStorage(row)">编辑</el-button><el-button v-if="!row.is_enabled" size="small" type="primary" @click="enableStorage(row)">启用</el-button><el-button v-if="!row.is_enabled" size="small" type="danger" @click="deleteStorage(row)">删除</el-button></template></el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane v-if="hasTab('resources')" label="资源策略" name="resources">
        <el-table :data="resourcePolicies">
          <el-table-column label="资源类型" width="120"><template #default="{ row }">{{ resourceTypeName(row.resource_type) }}</template></el-table-column>
          <el-table-column prop="key_prefix" label="对象路径前缀" min-width="160"/>
          <el-table-column label="保存天数" width="120"><template #default="{ row }">{{ row.retention_days === 0 ? '永久' : `${row.retention_days} 天` }}</template></el-table-column>
          <el-table-column label="公开访问" width="100"><template #default="{ row }"><el-tag :type="row.is_public ? 'success' : 'info'">{{ row.is_public ? '公开' : '私有' }}</el-tag></template></el-table-column>
          <el-table-column label="缓存" width="120"><template #default="{ row }">{{ row.cache_max_age }} 秒</template></el-table-column>
          <el-table-column label="文件上限" width="120"><template #default="{ row }">{{ row.max_size_mb }} MB</template></el-table-column>
          <el-table-column label="操作" width="90"><template #default="{ row }"><el-button link type="primary" @click="editResourcePolicy(row)">配置</el-button></template></el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="storageDialog" :title="storageForm.id ? '编辑存储配置' : '添加存储配置'" width="600px"><el-form :model="storageForm" label-width="110px"><el-form-item label="名称"><el-input v-model="storageForm.name" /></el-form-item><el-form-item label="类型"><el-select v-model="storageForm.type"><el-option label="本地存储" value="local"/><el-option label="腾讯云 COS" value="tencent_cos"/><el-option label="阿里云 OSS" value="aliyun_oss"/><el-option label="Cloudflare R2" value="cloudflare"/></el-select></el-form-item><el-form-item v-if="storageForm.type==='local'" label="本地路径"><el-input v-model="storageForm.local_path" /></el-form-item><template v-else><el-form-item label="Endpoint"><el-input v-model="storageForm.endpoint" /></el-form-item><el-form-item label="Access Key"><el-input v-model="storageForm.access_key" placeholder="编辑时留空表示不修改" /></el-form-item><el-form-item label="Secret Key"><el-input v-model="storageForm.secret_key" type="password" placeholder="编辑时留空表示不修改" /></el-form-item><el-form-item label="Bucket"><el-input v-model="storageForm.bucket" /></el-form-item><el-form-item label="Region"><el-input v-model="storageForm.region" /></el-form-item></template><el-form-item label="访问域名"><el-input v-model="storageForm.domain" /></el-form-item></el-form><template #footer><el-button @click="storageDialog=false">取消</el-button><el-button type="primary" @click="saveStorage">保存</el-button></template></el-dialog>
    <el-dialog v-model="resourcePolicyDialog" :title="`${resourceTypeName(resourcePolicyForm.resource_type)}策略`" width="560px"><el-form :model="resourcePolicyForm" label-width="130px"><el-form-item label="对象路径前缀"><el-input v-model="resourcePolicyForm.key_prefix" /></el-form-item><el-form-item label="保存天数"><el-input-number v-model="resourcePolicyForm.retention_days" :min="0" :max="3650" /><span class="field-tip">0 表示永久</span></el-form-item><el-form-item label="公开访问"><el-switch v-model="resourcePolicyForm.is_public" /></el-form-item><el-form-item label="缓存秒数"><el-input-number v-model="resourcePolicyForm.cache_max_age" :min="0" :max="31536000" /></el-form-item><el-form-item label="文件上限 MB"><el-input-number v-model="resourcePolicyForm.max_size_mb" :min="1" :max="10240" /></el-form-item></el-form><template #footer><el-button @click="resourcePolicyDialog=false">取消</el-button><el-button type="primary" @click="saveResourcePolicy">保存</el-button></template></el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const route = useRoute()
const configTabs = computed(() => route.meta.configTabs || ['basic', 'packages', 'email', 'task', 'prompt-optimization', 'storage', 'resources'])
const pageTitle = computed(() => route.meta.title || '系统配置')
const activeTab=ref(configTabs.value[0]),loading=ref(false),saving=ref(false),storageList=ref([]),storageDialog=ref(false),resourcePolicies=ref([]),resourcePolicyDialog=ref(false),allModels=ref([]),promptOptimizationLogs=ref([]),promptOptimizationLogTotal=ref(0),promptOptimizationLogPage=ref(1),creditPackages=ref([])
const hasTab = name => configTabs.value.includes(name)
const basicConfig=reactive({site_name:'',site_desc:'',register_enabled:true})
const emailConfig=reactive({smtp_host:'',smtp_port:587,smtp_user:'',smtp_password:'',smtp_ssl:false,from_name:'',from_email:'',is_active:false})
const taskConfig=reactive({max_active_tasks:50,prompt_max_length:5000,max_retry_attempts:0})
const promptOptimizationConfig=reactive({is_active:false,model_name:'',system_prompt:'',max_input_length:5000,credit_cost:0,rate_limit_per_minute:5})
const textModels=computed(()=>allModels.value.filter(model=>model.type==='text'))
const emptyStorage=()=>({id:0,name:'',type:'local',local_path:'./uploads',endpoint:'',access_key:'',secret_key:'',bucket:'',region:'',domain:''})
const storageForm=reactive(emptyStorage())
const resourcePolicyForm=reactive({resource_type:'',key_prefix:'',retention_days:0,is_public:true,cache_max_age:86400,max_size_mb:20})
const bool=v=>v===true||v==='true'

async function loadAll(){loading.value=true;try{const [basic,email,task,promptOptimization,storage,policies,models,packages]=await Promise.all([request.get('/config/basic'),request.get('/config/email'),request.get('/config/task'),request.get('/config/prompt-optimization'),request.get('/config/storage'),request.get('/config/storage/policies'),request.get('/config/models'),request.get('/config/credit-packages')]);Object.assign(basicConfig,{site_name:basic.site_name||'AGI Platform',site_desc:basic.site_desc||'',register_enabled:bool(basic.register_enabled)});Object.assign(emailConfig,email,{smtp_password:''});Object.assign(taskConfig,task);Object.assign(promptOptimizationConfig,promptOptimization);storageList.value=storage||[];resourcePolicies.value=policies||[];allModels.value=models||[];creditPackages.value=packages||[];await loadPromptOptimizationLogs(1)}finally{loading.value=false}}
async function saveBasicConfig(){saving.value=true;try{await request.put('/config/basic',basicConfig);ElMessage.success('基础配置已保存')}finally{saving.value=false}}
async function saveCreditPackages(){if(creditPackages.value.length!==3)return ElMessage.warning('必须保留三个套餐');if(creditPackages.value.some(item=>!item.name||!item.price||!item.points))return ElMessage.warning('请完善套餐名称、价格和灵感值');saving.value=true;try{await request.put('/config/credit-packages',creditPackages.value);ElMessage.success('充值套餐已保存');await loadAll()}finally{saving.value=false}}
async function saveEmailConfig(){saving.value=true;try{await request.put('/config/email',emailConfig);emailConfig.smtp_password='';ElMessage.success('邮件配置已保存')}finally{saving.value=false}}
async function saveTaskConfig(){saving.value=true;try{await request.put('/config/task',taskConfig);ElMessage.success('任务配置已保存')}finally{saving.value=false}}
async function savePromptOptimizationConfig(){saving.value=true;try{await request.put('/config/prompt-optimization',promptOptimizationConfig);ElMessage.success('提示词优化配置已保存')}finally{saving.value=false}}
async function loadPromptOptimizationLogs(page=promptOptimizationLogPage.value){promptOptimizationLogPage.value=page;const result=await request.get('/config/prompt-optimization/logs',{params:{page,page_size:20}});promptOptimizationLogs.value=result.list||[];promptOptimizationLogTotal.value=result.total||0}
async function testEmail(){const {value}=await ElMessageBox.prompt('请输入测试邮件接收地址','测试邮件',{inputPattern:/^[^\s@]+@[^\s@]+\.[^\s@]+$/,inputErrorMessage:'请输入有效邮箱'});await request.post('/config/email/test',{email:value});ElMessage.success('测试邮件已发送')}
function addStorage(){Object.assign(storageForm,emptyStorage());storageDialog.value=true}
function editStorage(row){Object.assign(storageForm,emptyStorage(),row,{access_key:'',secret_key:''});storageDialog.value=true}
async function saveStorage(){const body={...storageForm};delete body.id;if(storageForm.id)await request.put(`/config/storage/${storageForm.id}`,body);else await request.post('/config/storage',body);storageDialog.value=false;ElMessage.success('存储配置已保存');await loadAll()}
async function enableStorage(row){await request.post(`/config/storage/${row.id}/enable`);ElMessage.success('存储已启用');await loadAll()}
async function deleteStorage(row){await ElMessageBox.confirm(`确认删除“${row.name}”？`,'删除配置',{type:'warning'});await request.delete(`/config/storage/${row.id}`);ElMessage.success('删除成功');await loadAll()}
const getStorageTypeName=type=>({local:'本地存储',tencent_cos:'腾讯云COS',aliyun_oss:'阿里云OSS',cloudflare:'Cloudflare R2'}[type]||type)
const resourceTypeName=type=>({image:'生成图片',video:'生成视频',thumbnail:'视频缩略图',reference:'用户参考图',published_image:'发布图片',published_video:'发布视频',published_thumbnail:'发布视频封面'}[type]||type)
function editResourcePolicy(row){Object.assign(resourcePolicyForm,row);resourcePolicyDialog.value=true}
async function saveResourcePolicy(){await request.put(`/config/storage/policies/${resourcePolicyForm.resource_type}`,{key_prefix:resourcePolicyForm.key_prefix,retention_days:resourcePolicyForm.retention_days,is_public:resourcePolicyForm.is_public,cache_max_age:resourcePolicyForm.cache_max_age,max_size_mb:resourcePolicyForm.max_size_mb});resourcePolicyDialog.value=false;ElMessage.success('资源策略已保存');await loadAll()}
watch(configTabs, tabs => { if (!tabs.includes(activeTab.value)) activeTab.value = tabs[0] })
onMounted(loadAll)
</script>

<style scoped>.config-management{padding:20px}.config-form{max-width:760px}.mb{margin-bottom:20px}.mt{margin-top:16px}h2{margin-bottom:20px}.field-tip,.subtle,.package-tip{margin-left:10px;color:#909399;font-size:12px}.subtle{margin-left:0}.package-tip{margin:0 0 16px}.wide{width:100%}</style>
