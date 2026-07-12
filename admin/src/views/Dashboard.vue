<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon user">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-content">
              <p class="stat-label">总用户数</p>
              <h3 class="stat-value">{{ stats.total_users }}</h3>
              <p class="stat-sub">今日新增: {{ stats.today_users }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon task">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-content">
              <p class="stat-label">总任务数</p>
              <h3 class="stat-value">{{ stats.total_tasks }}</h3>
              <p class="stat-sub">今日新增: {{ stats.today_tasks }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon work">
              <el-icon><Picture /></el-icon>
            </div>
            <div class="stat-content">
              <p class="stat-label">总作品数</p>
              <h3 class="stat-value">{{ stats.total_works }}</h3>
              <p class="stat-sub">今日新增: {{ stats.today_works }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon pending">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="stat-content">
              <p class="stat-label">待审核作品</p>
              <h3 class="stat-value warning">{{ stats.pending_works }}</h3>
              <p class="stat-sub">需要处理</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getStats } from '@/api/admin'

const stats = ref({
  total_users: 0,
  total_tasks: 0,
  total_works: 0,
  pending_works: 0,
  today_users: 0,
  today_tasks: 0,
  today_works: 0
})

const loadStats = async () => {
  try {
    const data = await getStats()
    stats.value = data
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped lang="scss">
.dashboard {
  .stat-card {
    .stat-item {
      display: flex;
      gap: 15px;
    }

    .stat-icon {
      width: 60px;
      height: 60px;
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 28px;

      &.user {
        background: #ecf5ff;
        color: #409eff;
      }

      &.task {
        background: #f4f4f5;
        color: #909399;
      }

      &.work {
        background: #f0f9ff;
        color: #67c23a;
      }

      &.pending {
        background: #fef0f0;
        color: #f56c6c;
      }
    }

    .stat-content {
      flex: 1;

      .stat-label {
        margin: 0;
        color: #909399;
        font-size: 14px;
      }

      .stat-value {
        margin: 8px 0;
        font-size: 28px;
        font-weight: bold;

        &.warning {
          color: #f56c6c;
        }
      }

      .stat-sub {
        margin: 0;
        color: #c0c4cc;
        font-size: 13px;
      }
    }
  }
}
</style>
