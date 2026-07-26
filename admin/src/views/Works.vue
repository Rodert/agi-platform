<template>
  <div class="works-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <h3>作品管理</h3>
          <el-button @click="loadWorks" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-tabs v-model="status" @tab-change="handleStatusFilter">
        <el-tab-pane label="全部" name="" />
        <el-tab-pane label="待审核" name="pending" />
        <el-tab-pane label="已上架" name="approved" />
        <el-tab-pane label="已下架" name="offline" />
        <el-tab-pane label="已拒绝" name="rejected" />
      </el-tabs>

      <el-table :data="works" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />

        <el-table-column label="作品" width="120">
          <template #default="{ row }">
            <video v-if="row.type === 'video'" :src="row.video_url" :poster="row.image_url || undefined" muted preload="metadata" class="work-thumbnail" @click="openPreview(row)" />
            <el-image
              v-else
              :src="row.image_url"
              class="work-thumbnail"
              fit="cover"
              @click="openPreview(row)"
            />
          </template>
        </el-table-column>

        <el-table-column prop="title" label="标题" min-width="150" />
        <el-table-column prop="prompt" label="提示词" min-width="200" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="100" />

        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'image' ? 'success' : 'primary'">
              {{ row.type === 'image' ? '图片' : '视频' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="audit_status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.audit_status)">{{ statusLabel(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="提交时间" width="180" />

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <template v-if="row.audit_status === 'pending'">
              <el-button type="success" size="small" @click="handleAudit(row, 'approved')">通过</el-button>
              <el-button type="danger" size="small" @click="handleAudit(row, 'rejected')">拒绝</el-button>
            </template>
            <el-button v-else-if="row.audit_status === 'approved'" type="warning" size="small" @click="handlePublicationStatus(row, 'offline')">下架</el-button>
            <el-button v-else-if="row.audit_status === 'offline'" type="success" size="small" @click="handlePublicationStatus(row, 'approved')">重新上架</el-button>
            <span v-else class="muted">已拒绝</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadWorks"
          @size-change="loadWorks"
        />
      </div>
    </el-card>

    <el-dialog v-model="rejectDialogVisible" title="拒绝原因" width="500px">
      <el-form>
        <el-form-item label="拒绝原因">
          <el-input v-model="rejectReason" type="textarea" :rows="4" placeholder="请输入拒绝原因（选填）" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmReject">确定拒绝</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewVisible" title="作品预览" width="min(920px, calc(100vw - 32px))" destroy-on-close>
      <template v-if="previewWork">
        <div class="preview-copy">
          <h3>{{ previewWork.title || '未命名作品' }}</h3>
          <div><el-tag size="small">{{ previewWork.type === 'image' ? '图片' : '视频' }}</el-tag><el-tag v-if="previewWork.category" size="small" effect="plain">{{ previewWork.category }}</el-tag></div>
          <p>{{ previewWork.prompt || '未填写提示词' }}</p>
        </div>
        <video v-if="previewWork.type === 'video'" class="preview-media" :src="previewWork.video_url" :poster="previewWork.image_url || undefined" controls preload="metadata">当前浏览器不支持视频播放。</video>
        <img v-else class="preview-media" :src="previewWork.image_url" :alt="previewWork.title || '作品预览'" />
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getWorks, auditWork, updateWorkStatus } from '@/api/admin'

const loading = ref(false)
const works = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const status = ref('pending')
const rejectDialogVisible = ref(false)
const rejectReason = ref('')
const currentWork = ref(null)
const previewVisible = ref(false)
const previewWork = ref(null)

const loadWorks = async () => {
  loading.value = true
  try {
    const params = { page: currentPage.value, page_size: pageSize.value }
    if (status.value) params.status = status.value
    const data = await getWorks(params)
    works.value = data.list || data
    total.value = data.total || 0
  } catch (error) {
    console.error('加载作品失败:', error)
  } finally {
    loading.value = false
  }
}

const handleStatusFilter = () => {
  currentPage.value = 1
  loadWorks()
}

const statusLabel = (value) => ({ pending: '待审核', approved: '已上架', offline: '已下架', rejected: '已拒绝' }[value] || value)
const statusTagType = (value) => ({ pending: 'warning', approved: 'success', offline: 'info', rejected: 'danger' }[value] || 'info')
const openPreview = (work) => { previewWork.value = work; previewVisible.value = true }

const handleAudit = async (work, nextStatus) => {
  if (nextStatus === 'rejected') {
    currentWork.value = work
    rejectReason.value = ''
    rejectDialogVisible.value = true
    return
  }
  try {
    await ElMessageBox.confirm(`确定要通过“${work.title}”吗？`, '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'success' })
    await auditWork(work.id, 'approved')
    ElMessage.success('作品已上架到首页作品流')
    loadWorks()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') console.error('审核失败:', error)
  }
}

const confirmReject = async () => {
  try {
    await auditWork(currentWork.value.id, 'rejected', rejectReason.value)
    ElMessage.success('作品已拒绝')
    rejectDialogVisible.value = false
    loadWorks()
  } catch (error) {
    console.error('拒绝失败:', error)
  }
}

const handlePublicationStatus = async (work, nextStatus) => {
  const action = nextStatus === 'offline' ? '下架' : '重新上架'
  try {
    await ElMessageBox.confirm(`确定要${action}“${work.title}”吗？`, '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: nextStatus === 'offline' ? 'warning' : 'success' })
    await updateWorkStatus(work.id, nextStatus)
    ElMessage.success(`作品已${action}`)
    loadWorks()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') console.error(`${action}失败:`, error)
  }
}

onMounted(loadWorks)
</script>

<style scoped lang="scss">
.works-page {
  .card-header { display: flex; justify-content: space-between; align-items: center; h3 { margin: 0; } }
  .pagination { display: flex; justify-content: flex-end; margin-top: 20px; }
  .muted { color: #909399; font-size: 13px; }
  .work-thumbnail { width: 80px; height: 80px; border-radius: 4px; object-fit: cover; cursor: pointer; }
  .preview-copy { margin-bottom: 16px; .el-tag + .el-tag { margin-left: 8px; } h3 { margin: 0 0 8px; font-size: 18px; } p { margin: 10px 0 0; color: #606266; line-height: 1.7; white-space: pre-wrap; } }
  .preview-media { display: block; width: 100%; height: min(62vh, 620px); background: #111; border-radius: 4px; object-fit: contain; }
}
</style>
