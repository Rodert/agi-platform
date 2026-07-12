<template>
  <div class="works-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <h3>待审核作品</h3>
          <el-button @click="loadWorks" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table :data="works" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />

        <el-table-column label="作品" width="120">
          <template #default="{ row }">
            <el-image
              :src="row.image_url"
              :preview-src-list="[row.image_url]"
              style="width: 80px; height: 80px; border-radius: 4px"
              fit="cover"
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

        <el-table-column prop="created_at" label="创建时间" width="180" />

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="success"
              size="small"
              @click="handleAudit(row, 'approved')"
            >
              通过
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleAudit(row, 'rejected')"
            >
              拒绝
            </el-button>
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

    <!-- 拒绝原因对话框 -->
    <el-dialog v-model="rejectDialogVisible" title="拒绝原因" width="500px">
      <el-form>
        <el-form-item label="拒绝原因">
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="4"
            placeholder="请输入拒绝原因（选填）"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmReject">确定拒绝</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getPendingWorks, auditWork } from '@/api/admin'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const works = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const rejectDialogVisible = ref(false)
const rejectReason = ref('')
const currentWork = ref(null)

const loadWorks = async () => {
  loading.value = true
  try {
    const data = await getPendingWorks(currentPage.value, pageSize.value)
    works.value = data.list || data
    total.value = data.total || 0
  } catch (error) {
    console.error('加载作品失败:', error)
  } finally {
    loading.value = false
  }
}

const handleAudit = async (work, status) => {
  if (status === 'approved') {
    ElMessageBox.confirm('确定要通过这个作品吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'success'
    }).then(async () => {
      try {
        await auditWork(work.id, status)
        ElMessage.success('审核成功')
        loadWorks()
      } catch (error) {
        console.error('审核失败:', error)
      }
    })
  } else {
    currentWork.value = work
    rejectDialogVisible.value = true
    rejectReason.value = ''
  }
}

const confirmReject = async () => {
  try {
    await auditWork(currentWork.value.id, 'rejected', rejectReason.value)
    ElMessage.success('已拒绝')
    rejectDialogVisible.value = false
    loadWorks()
  } catch (error) {
    console.error('审核失败:', error)
  }
}

onMounted(() => {
  loadWorks()
})
</script>

<style scoped lang="scss">
.works-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
      margin: 0;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
