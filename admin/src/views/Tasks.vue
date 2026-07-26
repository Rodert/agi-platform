<template>
  <div class="tasks-page" v-loading="loading">
    <div class="page-head"><div><h2>生成记录</h2><p>所有用户的图片与视频生成任务，按创建时间倒排。</p></div></div>
    <el-form :inline="true" class="filters">
      <el-form-item label="用户/模型"><el-input v-model="filters.keyword" clearable placeholder="用户名、邮箱或模型名" @keyup.enter="search"/></el-form-item>
      <el-form-item label="类型"><el-select v-model="filters.type" clearable placeholder="全部" class="filter-select"><el-option label="图片" value="image"/><el-option label="视频" value="video"/></el-select></el-form-item>
      <el-form-item label="状态"><el-select v-model="filters.status" clearable placeholder="全部" class="filter-select"><el-option label="排队中" value="queued"/><el-option label="生成中" value="processing"/><el-option label="上传中" value="uploading"/><el-option label="成功" value="success"/><el-option label="失败" value="failed"/></el-select></el-form-item>
      <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item>
    </el-form>
    <el-table :data="tasks" class="task-table">
      <el-table-column label="预览" width="96"><template #default="{ row }"><el-image v-if="row.type === 'image' && (row.thumbnail_url || row.result_url)" :src="row.thumbnail_url || row.result_url" fit="cover" class="preview" :preview-src-list="[row.result_url || row.thumbnail_url]" preview-teleported/><el-image v-else-if="row.type === 'video' && videoThumbnail(row)" :src="videoThumbnail(row)" fit="cover" class="preview"/><span v-else-if="row.type === 'video'" class="empty-preview">视频</span><span v-else class="empty-preview">-</span></template></el-table-column>
      <el-table-column label="用户" min-width="160"><template #default="{ row }"><strong>{{ row.user_name || `用户 #${row.user_id}` }}</strong><small>{{ row.user_email }}</small></template></el-table-column>
      <el-table-column prop="model_name" label="模型" min-width="150"/>
      <el-table-column label="类型" width="80"><template #default="{ row }"><el-tag>{{ row.type === 'video' ? '视频' : '图片' }}</el-tag></template></el-table-column>
      <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag><small v-if="row.status === 'processing'">{{ row.progress }}%</small></template></el-table-column>
	  <el-table-column label="失败原因" min-width="200" show-overflow-tooltip><template #default="{ row }"><span v-if="row.status === 'failed'">{{ row.error_msg || '未返回具体原因' }}</span><span v-else>-</span></template></el-table-column>
	      <el-table-column label="重试" width="110"><template #default="{ row }"><span>{{ retryCount(row) }} / {{ row.max_retry_attempts }}</span><small>执行 {{ row.attempt_count }} 次</small></template></el-table-column>
      <el-table-column prop="cost" label="消耗" width="80"><template #default="{ row }">{{ row.cost }} 灵感值</template></el-table-column>
      <el-table-column prop="channel_name" label="渠道" min-width="130"/>
      <el-table-column prop="created_at" label="创建时间" width="170"/>
      <el-table-column label="操作" width="80" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openDetail(row)">详情</el-button></template></el-table-column>
    </el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[20,50,100]" layout="total, sizes, prev, pager, next" @size-change="load" @current-change="load" class="pager"/>
    <el-dialog v-model="detailVisible" title="生成任务详情" width="720px" @closed="selected = null"><template v-if="selected"><section v-if="selected.type === 'video' && selected.result_url" class="task-video"><h3>生成视频</h3><video :src="selected.result_url" :poster="videoThumbnail(selected) || undefined" controls preload="metadata">当前浏览器不支持视频播放。</video></section><el-descriptions :column="2" border><el-descriptions-item label="任务 ID">{{ selected.id }}</el-descriptions-item><el-descriptions-item label="状态">{{ statusText(selected.status) }}</el-descriptions-item><el-descriptions-item label="用户">{{ selected.user_name }} {{ selected.user_email }}</el-descriptions-item><el-descriptions-item label="渠道">{{ selected.channel_name || '-' }}</el-descriptions-item><el-descriptions-item label="模型">{{ selected.model_name }}</el-descriptions-item><el-descriptions-item label="消耗">{{ selected.cost }} 灵感值</el-descriptions-item><el-descriptions-item label="执行次数">{{ selected.attempt_count }} 次</el-descriptions-item><el-descriptions-item label="重试次数">{{ retryCount(selected) }} / {{ selected.max_retry_attempts }}</el-descriptions-item><el-descriptions-item v-if="selected.provider_task_id" label="上游任务 ID">{{ selected.provider_task_id }}</el-descriptions-item><el-descriptions-item v-if="selected.provider_status" label="上游状态">{{ selected.provider_status }}</el-descriptions-item><el-descriptions-item label="最近查询">{{ selected.last_polled_at || '-' }}</el-descriptions-item><el-descriptions-item label="最近重试">{{ selected.last_retry_at || '-' }}</el-descriptions-item><el-descriptions-item label="完成时间">{{ selected.completed_at || '-' }}</el-descriptions-item><el-descriptions-item label="提示词" :span="2">{{ selected.prompt }}</el-descriptions-item><el-descriptions-item label="生成参数" :span="2"><pre>{{ JSON.stringify(selected.params || {}, null, 2) }}</pre></el-descriptions-item><el-descriptions-item v-if="selected.provider_task_id" label="上游响应" :span="2"><pre>{{ JSON.stringify(selected.provider_response || {}, null, 2) }}</pre></el-descriptions-item><el-descriptions-item v-if="selected.error_msg" label="最终失败原因" :span="2">{{ selected.error_msg }}</el-descriptions-item></el-descriptions><section v-if="selected.assets && selected.assets.length" class="attempts"><h3>对象存储资源</h3><el-table :data="selected.assets" size="small"><el-table-column label="类型" width="100"><template #default="{ row }">{{ resourceTypeText(row.resource_type) }}</template></el-table-column><el-table-column prop="object_key" label="对象 Key" min-width="240" show-overflow-tooltip/><el-table-column prop="size_bytes" label="大小" width="110"><template #default="{ row }">{{ fileSize(row.size_bytes) }}</template></el-table-column><el-table-column prop="expires_at" label="到期时间" width="170"/></el-table></section><section v-if="selected.attempts && selected.attempts.length" class="attempts"><h3>执行记录</h3><el-timeline><el-timeline-item v-for="attempt in selected.attempts" :key="attempt.attempt" :timestamp="attempt.started_at" :type="attemptType(attempt.status)">{{ attempt.attempt }} 次执行：{{ attemptText(attempt.status) }}<small v-if="attempt.error_msg">{{ attempt.error_msg }}</small></el-timeline-item></el-timeline></section></template></el-dialog>
  </div>
</template>
<script setup>
import { onMounted, reactive, ref } from 'vue'
import request from '@/utils/request'
const loading = ref(false), tasks = ref([]), detailVisible = ref(false), selected = ref(null)
const filters = reactive({ keyword: '', type: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const statusText = value => ({ queued: '排队中', processing: '生成中', uploading: '上传中', success: '成功', failed: '失败' }[value] || value)
const statusType = value => ({ queued: 'info', processing: 'warning', uploading: 'warning', success: 'success', failed: 'danger' }[value] || 'info')
const retryCount = row => Math.max(0, Number(row.attempt_count || 0) - 1)
const attemptText = value => ({ processing: '执行中', success: '成功', failed: '失败' }[value] || value)
const attemptType = value => ({ processing: 'warning', success: 'success', failed: 'danger' }[value] || 'info')
const resourceTypeText = value => ({image:'生成图片',video:'生成视频',thumbnail:'缩略图',reference:'参考图'}[value] || value)
const fileSize = value => value >= 1024 * 1024 ? `${(value / 1024 / 1024).toFixed(2)} MB` : `${Math.ceil(value / 1024)} KB`
const videoThumbnail = task => task.thumbnail_url && task.thumbnail_url !== task.result_url ? task.thumbnail_url : ''
async function load(){ loading.value=true; try { const data = await request.get('/tasks',{params:{...filters,page:pagination.page,page_size:pagination.pageSize}}); tasks.value=data.list||[]; pagination.total=data.total||0 } finally { loading.value=false } }
function search(){ pagination.page=1; load() }
function reset(){ Object.assign(filters,{keyword:'',type:'',status:''}); search() }
function openDetail(row){ selected.value=row; detailVisible.value=true }
onMounted(load)
</script>
<style scoped>.tasks-page{padding:20px}.page-head h2{margin:0}.page-head p{color:#909399;font-size:13px}.filters{margin:20px 0}.filter-select{width:120px}.preview{width:64px;height:64px;border-radius:4px;background:#f5f7fa;object-fit:cover}.empty-preview{display:inline-grid;width:64px;height:64px;place-items:center;color:#c0c4cc;background:#f5f7fa}small{display:block;color:#909399;margin-top:3px}.pager{margin-top:20px;justify-content:flex-end}pre{white-space:pre-wrap;margin:0;font-size:12px}.attempts{margin-top:20px}.attempts h3,.task-video h3{font-size:14px;margin:0 0 14px}.task-video{margin-bottom:20px}.task-video video{display:block;width:100%;max-height:420px;background:#111;border-radius:4px}</style>
