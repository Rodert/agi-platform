<template>
  <div class="reports" v-loading="loading">
    <div class="page-header">
      <div>
        <h2>数据报表</h2>
        <p>用户、生成、积分消耗与内容运营数据</p>
      </div>
      <el-date-picker
        v-model="dateRange"
        class="report-date-picker"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        format="YYYY-MM-DD"
        unlink-panels
        teleported
        popper-class="report-date-popper"
        :shortcuts="dateShortcuts"
        :clearable="false"
        :disabled-date="disableFutureDate"
        @change="loadReport"
      />
    </div>

    <el-row :gutter="16" class="summary-grid">
      <el-col v-for="item in summaryCards" :key="item.label" :xs="12" :sm="8" :lg="4">
        <el-card shadow="never" class="summary-card">
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
          <div class="summary-note">{{ item.note }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :lg="14">
        <el-card shadow="never">
          <template #header><span>用户与任务趋势</span></template>
          <div ref="trendChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="10">
        <el-card shadow="never">
          <template #header><span>任务状态</span></template>
          <div ref="taskStatusChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :lg="12">
        <el-card shadow="never">
          <template #header><span>积分消耗趋势</span></template>
          <div ref="creditsChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card shadow="never">
          <template #header><span>作品运营</span></template>
          <div ref="worksChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :lg="8">
        <el-card shadow="never">
          <template #header><span>任务类型</span></template>
          <div ref="taskTypesChart" class="chart small-chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="8">
        <el-card shadow="never">
          <template #header><span>模型用量</span></template>
          <div ref="modelsChart" class="chart small-chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="8">
        <el-card shadow="never">
          <template #header><span>渠道用量</span></template>
          <div ref="channelsChart" class="chart small-chart"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { getReport } from '@/api/admin'

const loading = ref(false)
const report = ref(emptyReport())
const trendChart = ref(null)
const taskStatusChart = ref(null)
const creditsChart = ref(null)
const worksChart = ref(null)
const taskTypesChart = ref(null)
const modelsChart = ref(null)
const channelsChart = ref(null)
const charts = []

const formatDate = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const end = new Date()
const start = new Date()
start.setDate(start.getDate() - 29)
const dateRange = ref([formatDate(start), formatDate(end)])

const dateShortcuts = [
  { text: '今天', value: () => [new Date(), new Date()] },
  { text: '近 7 天', value: () => rangeFromToday(6) },
  { text: '近 30 天', value: () => rangeFromToday(29) }
]

function rangeFromToday (daysAgo) {
  const rangeEnd = new Date()
  const rangeStart = new Date()
  rangeStart.setDate(rangeStart.getDate() - daysAgo)
  return [rangeStart, rangeEnd]
}

function emptyReport () {
  return {
    summary: { new_users: 0, active_users: 0, tasks: 0, success_tasks: 0, failed_tasks: 0, pending_tasks: 0, success_rate: 0, credits_consumed: 0, works: 0, pending_works: 0, approved_works: 0, rejected_works: 0, offline_works: 0 },
    daily: [], task_types: [], task_models: [], task_channels: [], work_statuses: [], work_categories: []
  }
}

const summaryCards = computed(() => {
  const data = report.value.summary
  return [
    { label: '新增用户', value: data.new_users, note: `活跃用户 ${data.active_users}` },
    { label: '生成任务', value: data.tasks, note: `成功率 ${data.success_rate.toFixed(1)}%` },
    { label: '成功任务', value: data.success_tasks, note: `失败 ${data.failed_tasks}，处理中 ${data.pending_tasks}` },
    { label: '积分消耗', value: data.credits_consumed, note: '以实际扣减流水统计' },
    { label: '新增作品', value: data.works, note: `待审核 ${data.pending_works}` },
    { label: '审核通过', value: data.approved_works, note: `驳回 ${data.rejected_works}，下架 ${data.offline_works}` }
  ]
})

const disableFutureDate = (date) => date.getTime() > Date.now()
const palette = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#9b7bdb']

function makeChart (element) {
  const chart = echarts.init(element)
  charts.push(chart)
  return chart
}

function pieOption (data, emptyText = '暂无数据') {
  return {
    color: palette,
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    graphic: data.length ? [] : [{ type: 'text', left: 'center', top: 'middle', style: { text: emptyText, fill: '#909399', fontSize: 14 } }],
    series: [{ type: 'pie', radius: ['42%', '70%'], avoidLabelOverlap: true, label: { formatter: '{b}\n{c}' }, data }]
  }
}

function renderCharts () {
  const data = report.value
  const dates = data.daily.map(item => item.date.slice(5))
  charts.forEach(chart => chart.dispose())
  charts.length = 0

  if (![trendChart, taskStatusChart, creditsChart, worksChart, taskTypesChart, modelsChart, channelsChart].every(chart => chart.value)) {
    return
  }

  makeChart(trendChart.value).setOption({
    color: palette,
    tooltip: { trigger: 'axis' },
    legend: { data: ['新增用户', '活跃用户', '生成任务', '新增作品'] },
    grid: { left: 44, right: 20, top: 42, bottom: 30 },
    xAxis: { type: 'category', boundaryGap: false, data: dates },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      { name: '新增用户', type: 'line', smooth: true, data: data.daily.map(item => item.new_users) },
      { name: '活跃用户', type: 'line', smooth: true, data: data.daily.map(item => item.active_users) },
      { name: '生成任务', type: 'line', smooth: true, data: data.daily.map(item => item.tasks) },
      { name: '新增作品', type: 'line', smooth: true, data: data.daily.map(item => item.works) }
    ]
  })
  makeChart(taskStatusChart.value).setOption(pieOption([
    { name: '成功', value: data.summary.success_tasks },
    { name: '失败', value: data.summary.failed_tasks },
    { name: '处理中', value: data.summary.pending_tasks }
  ].filter(item => item.value > 0)))
  makeChart(creditsChart.value).setOption({
    color: [palette[2]], tooltip: { trigger: 'axis' }, grid: { left: 48, right: 20, top: 20, bottom: 30 },
    xAxis: { type: 'category', data: dates }, yAxis: { type: 'value', minInterval: 1 },
    series: [{ name: '积分消耗', type: 'bar', data: data.daily.map(item => item.credits_consumed), barMaxWidth: 30 }]
  })
  makeChart(worksChart.value).setOption({
    color: palette, tooltip: { trigger: 'axis' },
    grid: { left: 92, right: 25, top: 14, bottom: 20 },
    xAxis: { type: 'value', minInterval: 1 },
    yAxis: { type: 'category', data: [
      ...data.work_categories.map(item => `分类 · ${item.name}`),
      ...data.work_statuses.map(item => `审核 · ${item.name}`)
    ] },
    series: [{
      name: '作品数', type: 'bar', barMaxWidth: 26,
      data: [...data.work_categories, ...data.work_statuses].map(item => item.value)
    }]
  })
  makeChart(taskTypesChart.value).setOption(pieOption(data.task_types))
  makeChart(modelsChart.value).setOption(pieOption(data.task_models))
  makeChart(channelsChart.value).setOption(pieOption(data.task_channels))
}

async function loadReport () {
  if (!dateRange.value || dateRange.value.length !== 2) return
  loading.value = true
  try {
    const data = await getReport(dateRange.value[0], dateRange.value[1])
    report.value = { ...emptyReport(), ...data, summary: { ...emptyReport().summary, ...data.summary } }
  } catch (error) {
    report.value = emptyReport()
    ElMessage.error(error.message || '加载数据报表失败，请稍后重试')
  } finally {
    loading.value = false
  }
  await nextTick()
  try {
    renderCharts()
  } catch (error) {
    console.error('报表图表渲染失败：', error)
    ElMessage.warning('数据已加载，但部分图表暂未显示')
  }
}

const resizeCharts = () => charts.forEach(chart => chart.resize())

onMounted(() => {
  loadReport()
  window.addEventListener('resize', resizeCharts)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  charts.forEach(chart => chart.dispose())
})
</script>

<style scoped lang="scss">
.reports { padding: 0; }
.page-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.page-header h2 { margin: 0; font-size: 20px; color: #303133; }
.page-header p { margin: 6px 0 0; color: #909399; font-size: 13px; }
.report-date-picker { width: 270px; }
.summary-grid { margin-bottom: 16px; }
.summary-grid :deep(.el-col) { margin-bottom: 16px; }
.summary-card { min-height: 125px; }
.summary-label, .summary-note { color: #909399; font-size: 13px; }
.summary-value { color: #303133; font-size: 28px; font-weight: 600; line-height: 1.5; margin: 8px 0; }
.chart-row { margin-bottom: 16px; }
.chart { height: 300px; }
.small-chart { height: 260px; }
:global(.report-date-popper) { z-index: 3000 !important; }
@media (max-width: 768px) {
  .page-header { align-items: flex-start; flex-direction: column; }
  .report-date-picker { width: 100%; }
}
</style>
