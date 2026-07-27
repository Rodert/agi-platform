<template>
  <div class="channels-page" v-loading="loading">
    <div class="page-head">
      <div><h2>渠道与模型</h2><p>每行一个渠道账号；模型能力在全局目录统一维护。</p></div>
      <div><el-button plain @click="purgeUnreferencedModels">清理未引用模型</el-button><el-button type="primary" @click="openChannel()">添加渠道</el-button></div>
    </div>

    <div v-if="activeChannel" class="active-account" aria-live="polite">
      <span class="active-account-label">当前操作账号</span>
      <strong>{{ activeChannel.name || '未命名账号' }}</strong>
      <el-tag size="small" effect="plain">{{ activeChannel.provider }}</el-tag>
      <span class="active-account-url">{{ activeChannel.api_url || '未设置接口地址' }}</span>
      <el-button link type="primary" @click="openChannel(activeChannel)">编辑此账号</el-button>
    </div>

    <el-table :data="channels" row-key="id" class="channel-table" :row-class-name="channelRowClass" @row-click="selectChannel">
      <el-table-column type="expand" width="52">
        <template #default="{ row }">
          <div class="model-list">
            <div class="model-list-head">
              <strong>已识别模型</strong>
              <div>
                <el-button size="small" :loading="syncing === row.id" @click.stop="syncModels(row)">重新同步上游模型</el-button>
                <el-button size="small" type="primary" plain @click.stop="openBind(row)">手动绑定模型</el-button>
                <el-button v-if="selectedModelIDs(row).length" size="small" type="danger" plain @click.stop="removeSelectedModels(row)">删除选中 ({{ selectedModelIDs(row).length }})</el-button>
              </div>
            </div>
            <el-table :data="row.channel_models || []" size="small" empty-text="尚未识别到模型" @selection-change="items => setSelectedModels(row.id, items)">
              <el-table-column type="selection" width="46"/>
              <el-table-column label="模型名" min-width="180"><template #default="{ row: binding }"><code>{{ binding.model?.name }}</code><span class="display-name">{{ binding.model?.display_name }}</span></template></el-table-column>
              <el-table-column label="类型" width="90"><template #default="{ row: binding }"><el-tag>{{ ({ image: '图片', video: '视频', text: '文本' })[binding.model?.type] || binding.model?.type }}</el-tag></template></el-table-column>
              <el-table-column label="能力" min-width="230"><template #default="{ row: binding }">{{ capabilitySummary(binding.model) }}</template></el-table-column>
              <el-table-column label="启用" width="85"><template #default="{ row: binding }"><el-switch :model-value="binding.is_active" @change="value => setBindingStatus(row, binding, value)"/></template></el-table-column>
              <el-table-column label="操作" width="150"><template #default="{ row: binding }"><el-button link type="primary" @click.stop="openModel(binding.model, row)">配置能力</el-button><el-button link type="danger" @click.stop="removeModel(row, binding)">删除</el-button></template></el-table-column>
            </el-table>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="渠道账号" min-width="160"/>
      <el-table-column prop="provider" label="渠道" width="120"/>
      <el-table-column prop="api_url" label="接口地址" min-width="260" show-overflow-tooltip/>
      <el-table-column prop="priority" label="优先级" width="100"><template #default="{ row }">{{ row.priority }}</template></el-table-column>
      <el-table-column label="健康状态" width="100"><template #default="{ row }"><el-tag :type="healthType(row.health_status)">{{ healthLabel(row.health_status) }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.is_active ? 'success' : 'info'">{{ row.is_active ? '启用' : '停用' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="150" fixed="right"><template #default="{ row }"><el-button link type="primary" @click.stop="openChannel(row)">编辑</el-button><el-button link type="danger" @click.stop="removeChannel(row)">删除</el-button></template></el-table-column>
    </el-table>

    <el-dialog v-model="channelDialog" :title="channelForm.id ? `编辑账号：${channelForm.name || '未命名账号'}` : '添加渠道账号'" width="600px">
      <div v-if="channelForm.id" class="dialog-context"><span>正在编辑</span><strong>{{ channelForm.name || '未命名账号' }}</strong><el-tag size="small" effect="plain">{{ channelForm.provider }}</el-tag><span>{{ channelForm.api_url || '未设置接口地址' }}</span></div>
      <el-form :model="channelForm" label-width="100px">
        <el-form-item label="渠道账号"><el-input v-model="channelForm.name" placeholder="例如 ChatGPT 主账号"/></el-form-item>
        <el-form-item label="渠道"><el-select v-model="channelForm.provider" class="full" @change="ensureProviderConfig"><el-option v-for="item in providers" :key="item" :value="item" :label="item"/></el-select></el-form-item>
        <el-form-item label="API 地址"><el-input v-model="channelForm.api_url"/></el-form-item>
        <el-form-item label="API Key"><el-input v-model="channelForm.api_key" type="password" show-password :placeholder="channelForm.id ? '留空表示不修改' : '请输入 API Key'"/></el-form-item>
        <template v-if="channelForm.provider === 'grok'"><el-divider>Grok 接口配置</el-divider><el-form-item label="模型列表路径"><el-input v-model="channelForm.extra_config.models_path" placeholder="/v1/models"/></el-form-item><el-form-item label="创建任务路径"><el-input v-model="channelForm.extra_config.create_path" placeholder="/v1/video/generations"/></el-form-item><el-form-item label="查询状态路径"><el-input v-model="channelForm.extra_config.status_path" placeholder="/v1/video/generations/{task_id}"/></el-form-item><el-form-item label="参考图字段"><el-input v-model="channelForm.extra_config.reference_field" placeholder="images"/></el-form-item><el-form-item label="轮询间隔"><el-input-number v-model="channelForm.extra_config.poll_interval_seconds" :min="1" :max="60"/><span class="hint">秒</span></el-form-item><el-form-item label="轮询超时"><el-input-number v-model="channelForm.extra_config.poll_timeout_seconds" :min="30" :max="3600"/><span class="hint">秒</span></el-form-item></template>
        <template v-if="channelForm.provider === 'jimeng_international'"><el-divider>即梦国际接口配置</el-divider><el-form-item label="模型列表路径"><el-input v-model="channelForm.extra_config.models_path" placeholder="/v1/models"/></el-form-item><el-form-item label="创建任务路径"><el-input v-model="channelForm.extra_config.create_path" placeholder="/videos"/></el-form-item><el-form-item label="查询状态路径"><el-input v-model="channelForm.extra_config.status_path" placeholder="/videos/{task_id}"/></el-form-item><el-form-item label="成品下载路径"><el-input v-model="channelForm.extra_config.content_path" placeholder="/videos/{task_id}/content"/></el-form-item><el-form-item label="轮询间隔"><el-input-number v-model="channelForm.extra_config.poll_interval_seconds" :min="1" :max="60"/><span class="hint">秒</span></el-form-item><el-form-item label="轮询超时"><el-input-number v-model="channelForm.extra_config.poll_timeout_seconds" :min="30" :max="3600"/><span class="hint">秒</span></el-form-item></template>
        <el-form-item label="优先级"><el-input-number v-model="channelForm.priority" :min="1" :max="9999"/><span class="hint">数字越小越优先</span></el-form-item>
        <el-form-item label="启用"><el-switch v-model="channelForm.is_active"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="channelDialog = false">取消</el-button><el-button type="primary" @click="saveChannel">保存</el-button></template>
    </el-dialog>

    <el-dialog v-model="bindDialog" :title="`为 ${bindTargetName} 绑定模型`" width="500px">
      <div class="dialog-context"><span>目标账号</span><strong>{{ bindTargetName }}</strong></div>
      <el-form :model="bindForm" label-width="95px">
        <el-form-item label="模型名"><el-input v-model="bindForm.model_name" placeholder="例如 gpt-image-2"/></el-form-item>
        <el-form-item label="类型"><el-radio-group v-model="bindForm.type"><el-radio value="image">图片</el-radio><el-radio value="video">视频</el-radio><el-radio value="text">文本</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="bindDialog = false">取消</el-button><el-button type="primary" @click="bindModel">绑定</el-button></template>
    </el-dialog>

    <el-dialog v-model="modelDialog" :title="`模型能力配置：${modelForm.name}`" width="680px">
      <el-alert type="info" :closable="false" show-icon :title="modelContextName ? `正在从账号「${modelContextName}」查看；该配置属于全局模型，所有绑定渠道和用户端都会使用相同的能力集合。` : '该配置属于全局模型。所有已绑定渠道和用户端都会使用相同的能力集合。'" class="mb-5"/>
      <el-form :model="modelForm" label-width="110px">
        <el-form-item label="模型名"><el-input v-model="modelForm.name" disabled/></el-form-item>
        <el-form-item label="用户端名称"><el-input v-model="modelForm.display_name"/></el-form-item>
        <el-form-item label="基础积分"><el-input-number v-model="modelForm.cost" :min="0"/></el-form-item>
        <el-form-item label="描述"><el-input v-model="modelForm.description" type="textarea"/></el-form-item>
        <el-divider>支持能力</el-divider>
        <el-form-item label="画面比例"><div class="option-editor"><div v-for="(option, index) in modelForm.ratio_options" :key="`ratio-${index}`" class="option-row"><el-select v-model="option.value" filterable allow-create default-first-option placeholder="选择或输入比例" class="option-value"><el-option v-for="item in ratioPresets" :key="item" :label="item" :value="item"/></el-select><el-tooltip content="删除比例"><el-button circle text type="danger" :icon="Delete" @click="removeOption(modelForm.ratio_options, index)"/></el-tooltip></div><el-button plain :icon="Plus" @click="addOption(modelForm.ratio_options)">添加比例</el-button></div></el-form-item>
        <el-form-item label="清晰度"><div class="option-editor"><div v-for="(option, index) in modelForm.resolution_options" :key="`resolution-${index}`" class="option-row"><el-select v-model="option.value" filterable allow-create default-first-option placeholder="选择或输入清晰度" class="option-value"><el-option v-for="item in resolutionPresets" :key="item" :label="item" :value="item"/></el-select><el-input-number v-model="option.extra_cost" :min="0" controls-position="right" class="option-cost"/><span class="cost-label">附加积分</span><el-tooltip content="删除清晰度"><el-button circle text type="danger" :icon="Delete" @click="removeOption(modelForm.resolution_options, index)"/></el-tooltip></div><el-button plain :icon="Plus" @click="addOption(modelForm.resolution_options)">添加清晰度</el-button></div></el-form-item>
        <template v-if="modelForm.type === 'video'"><el-form-item label="视频时长"><div class="option-editor"><div v-for="(option, index) in modelForm.duration_options" :key="`duration-${index}`" class="option-row"><el-select v-model="option.value" filterable allow-create default-first-option placeholder="选择或输入时长" class="option-value"><el-option v-for="item in durationPresets" :key="item" :label="item" :value="item"/></el-select><el-input-number v-model="option.extra_cost" :min="0" controls-position="right" class="option-cost"/><span class="cost-label">附加积分</span><el-tooltip content="删除时长"><el-button circle text type="danger" :icon="Delete" @click="removeOption(modelForm.duration_options, index)"/></el-tooltip></div><el-button plain :icon="Plus" @click="addOption(modelForm.duration_options)">添加时长</el-button></div></el-form-item><el-form-item label="支持声音"><el-switch v-model="modelForm.sound_enabled"/><span class="hint">开启后用户端显示声音开关</span></el-form-item><el-form-item v-if="modelForm.sound_enabled" label="默认声音"><el-switch v-model="modelForm.sound_default"/></el-form-item></template>
      </el-form>
      <template #footer><el-button @click="modelDialog = false">取消</el-button><el-button type="primary" @click="saveModel">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'

const loading = ref(false)
const syncing = ref(0)
const channels = ref([])
const activeChannelId = ref(0)
const channelDialog = ref(false)
const bindDialog = ref(false)
const modelDialog = ref(false)
const providers = ['openai', 'chatgpt', 'gemini', 'grok', 'jimeng', 'jimeng_international', 'wave', 'demo']
const emptyChannel = () => ({ id: 0, name: '', provider: 'openai', api_url: '', api_key: '', is_active: true, priority: 100, extra_config: {} })
const channelForm = reactive(emptyChannel())
const bindForm = reactive({ channel_id: 0, model_name: '', type: 'image' })
const modelForm = reactive({ id: 0, name: '', type: 'image', display_name: '', cost: 0, description: '', ratio_options: [], resolution_options: [], duration_options: [], sound_enabled: false, sound_default: false })
const modelContextName = ref('')
const selectedModels = reactive({})
const ratioPresets = ['1:1', '4:5', '3:4', '16:9', '9:16', '21:9']
const resolutionPresets = ['480P', '720P', '1080P', '1K', '2K', '4K']
const durationPresets = ['3s', '5s', '8s', '10s', '15s']
const activeChannel = computed(() => channels.value.find(channel => channel.id === activeChannelId.value) || null)
const bindTargetName = computed(() => channels.value.find(channel => channel.id === bindForm.channel_id)?.name || '该账号')
const selectedModelIDs = channel => (selectedModels[channel.id] || []).map(binding => binding.model_id)

const load = async () => { loading.value = true; try { channels.value = await request.get('/channels'); if (activeChannelId.value && !channels.value.some(channel => channel.id === activeChannelId.value)) activeChannelId.value = 0 } finally { loading.value = false } }
const healthLabel = value => ({ healthy: '正常', unhealthy: '异常', unknown: '未检测' }[value] || '未检测')
const healthType = value => ({ healthy: 'success', unhealthy: 'danger', unknown: 'info' }[value] || 'info')
const optionText = config => (config?.options || []).map(item => item.value).join('、')
const capabilitySummary = model => { const params = model?.params_config || {}; return [params.ratio && `比例 ${optionText(params.ratio)}`, params.resolution && `清晰度 ${optionText(params.resolution)}`, params.duration && `时长 ${optionText(params.duration)}`, params.sound && '声音'].filter(Boolean).join(' · ') || '待配置' }
const optionRows = config => (config?.options || []).map(item => ({ value: item.value, extra_cost: item.extra_cost || 0 }))
const makeSelect = (label, rows) => { const options = rows.filter(item => item.value?.trim()).map(item => { const option = { value: item.value.trim(), label: item.value.trim() }; if (Number(item.extra_cost) > 0) option.extra_cost = Number(item.extra_cost); return option }); return options.length ? { label, type: 'select', default: options[0].value, options } : null }
const addOption = options => options.push({ value: '', extra_cost: 0 })
const removeOption = (options, index) => options.splice(index, 1)

function ensureProviderConfig() {
  if (!channelForm.extra_config || typeof channelForm.extra_config !== 'object') channelForm.extra_config = {}
  if (channelForm.provider === 'grok') Object.assign(channelForm.extra_config, { models_path: channelForm.extra_config.models_path || '/v1/models', create_path: channelForm.extra_config.create_path || '/v1/video/generations', status_path: channelForm.extra_config.status_path || '/v1/video/generations/{task_id}', reference_field: channelForm.extra_config.reference_field || 'images', poll_interval_seconds: channelForm.extra_config.poll_interval_seconds || 5, poll_timeout_seconds: channelForm.extra_config.poll_timeout_seconds || 900 })
  if (channelForm.provider === 'jimeng_international') Object.assign(channelForm.extra_config, { models_path: channelForm.extra_config.models_path || '/v1/models', create_path: channelForm.extra_config.create_path || '/videos', status_path: channelForm.extra_config.status_path || '/videos/{task_id}', content_path: channelForm.extra_config.content_path || '/videos/{task_id}/content', poll_interval_seconds: channelForm.extra_config.poll_interval_seconds || 5, poll_timeout_seconds: channelForm.extra_config.poll_timeout_seconds || 900 })
}
function selectChannel(row) { activeChannelId.value = row.id }
function channelRowClass({ row }) { return row.id === activeChannelId.value ? 'is-active-account' : '' }
function openChannel(row) { if (row) selectChannel(row); Object.assign(channelForm, emptyChannel(), row || {}, { api_key: '' }); ensureProviderConfig(); channelDialog.value = true }
async function saveChannel() { const body = { ...channelForm }; delete body.id; if (channelForm.id) await request.put(`/channels/${channelForm.id}`, body); else await request.post('/channels', body); channelDialog.value = false; ElMessage.success('渠道已保存'); await load() }
async function removeChannel(row) { selectChannel(row); await ElMessageBox.confirm(`将删除账号「${row.name}」（${row.provider}）及其全部模型绑定。此操作无法撤销。`, '确认删除当前账号', { type: 'warning', confirmButtonText: `删除「${row.name}」`, cancelButtonText: '取消' }); await request.delete(`/channels/${row.id}`); ElMessage.success(`账号「${row.name}」已删除`); await load() }
async function syncModels(row) { selectChannel(row); await ElMessageBox.confirm(`将以账号「${row.name}」当前上游模型列表为准，解除上游已不再返回的模型绑定。历史任务不会删除。`, '重新同步上游模型', { type: 'warning', confirmButtonText: '重新同步' }); syncing.value = row.id; try { const result = await request.post(`/channels/${row.id}/sync-models`); ElMessage.success(`账号「${row.name}」已同步 ${result.length} 个模型`); await load() } finally { syncing.value = 0 } }
async function purgeUnreferencedModels() { await ElMessageBox.confirm('将删除没有任何渠道绑定、且没有排队或生成中任务引用的全局模型。历史任务和作品不会删除。', '清理未引用模型', { type: 'warning', confirmButtonText: '清理' }); const result = await request.delete('/config/models/unreferenced'); ElMessage.success(`已清理 ${result.count || 0} 个未引用模型`); await load() }
function openBind(row) { selectChannel(row); Object.assign(bindForm, { channel_id: row.id, model_name: '', type: 'image' }); bindDialog.value = true }
async function bindModel() { await request.post(`/channels/${bindForm.channel_id}/models`, { model_name: bindForm.model_name, type: bindForm.type, is_active: true }); bindDialog.value = false; ElMessage.success('模型已绑定'); await load() }
async function setBindingStatus(channel, binding, value) { selectChannel(channel); await request.put(`/channels/${channel.id}/models/${binding.model_id}/status`, { is_active: value }); binding.is_active = value; ElMessage.success(`账号「${channel.name}」中的模型「${binding.model?.name}」已${value ? '启用' : '停用'}`) }
function setSelectedModels(channelID, items) { selectedModels[channelID] = items }
async function removeModel(channel, binding) { await ElMessageBox.confirm(`仅解除「${binding.model?.name}」与账号「${channel.name}」的绑定，不会删除全局模型。`, '确认删除模型绑定', { type: 'warning' }); await request.delete(`/channels/${channel.id}/models/${binding.model_id}`); ElMessage.success('模型绑定已删除'); selectedModels[channel.id] = []; await load() }
async function removeSelectedModels(channel) { const ids = selectedModelIDs(channel); if (!ids.length) return; await ElMessageBox.confirm(`将解除选中的 ${ids.length} 个模型绑定，不会删除全局模型。`, '确认批量删除', { type: 'warning' }); await request.delete(`/channels/${channel.id}/models`, { data: { model_ids: ids } }); ElMessage.success(`已删除 ${ids.length} 个模型绑定`); selectedModels[channel.id] = []; await load() }
function openModel(model, channel) { if (channel) selectChannel(channel); modelContextName.value = channel?.name || ''; const params = model.params_config || {}; Object.assign(modelForm, { id: model.id, name: model.name, type: model.type, display_name: model.display_name, cost: model.cost, description: model.description || '', ratio_options: optionRows(params.ratio), resolution_options: optionRows(params.resolution), duration_options: optionRows(params.duration), sound_enabled: Boolean(params.sound), sound_default: Boolean(params.sound?.default) }); modelDialog.value = true }
async function saveModel() { const params = {}; const ratio = makeSelect('画面比例', modelForm.ratio_options); const resolution = makeSelect('清晰度', modelForm.resolution_options); if (ratio) params.ratio = ratio; if (resolution) params.resolution = resolution; if (modelForm.type === 'video') { const duration = makeSelect('时长', modelForm.duration_options); if (duration) params.duration = duration; if (modelForm.sound_enabled) params.sound = { label: '生成声音', type: 'switch', default: modelForm.sound_default } } await request.put(`/config/models/${modelForm.id}`, { display_name: modelForm.display_name, cost: modelForm.cost, description: modelForm.description, params_config: params }); modelDialog.value = false; ElMessage.success('模型能力已保存'); await load() }

onMounted(load)
</script>

<style scoped>
.channels-page { padding: 20px; }
.page-head, .model-list-head { display: flex; justify-content: space-between; align-items: center; }
.page-head { margin-bottom: 20px; }.page-head h2 { margin: 0; }.page-head p, .hint, .cost-label { color: #909399; font-size: 13px; }.full { width: 100%; }.channel-table { background: #fff; }.channel-table :deep(.is-active-account > td.el-table__cell) { background: #ecf5ff !important; }.channel-table :deep(.is-active-account > td:first-child) { box-shadow: inset 3px 0 0 #409eff; }.active-account, .dialog-context { display: flex; align-items: center; gap: 9px; }.active-account { min-height: 48px; margin-bottom: 12px; padding: 10px 14px; background: #ecf5ff; border: 1px solid #b3d8ff; border-radius: 4px; color: #303133; }.active-account-label, .dialog-context > span:first-child { color: #606266; font-size: 13px; }.active-account-url { flex: 1; min-width: 0; overflow: hidden; color: #606266; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }.dialog-context { margin: -4px 0 18px; padding: 10px 12px; background: #f5f7fa; border-radius: 4px; color: #606266; font-size: 13px; }.model-list { padding: 10px 22px 18px; background: #f8fafc; }.model-list-head { margin-bottom: 12px; }.display-name { margin-left: 9px; color: #909399; }.mb-5 { margin-bottom: 20px; }.option-editor { width: 100%; }.option-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }.option-value { width: 220px; }.option-cost { width: 120px; }
</style>
