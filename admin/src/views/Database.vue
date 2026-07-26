<template>
  <div class="database-page">
    <el-card class="database-card">
      <template #header><div class="card-header"><div><h3>数据管理</h3><p>只读浏览当前数据库中的表与记录</p></div><el-button :loading="tableLoading" @click="loadTables"><el-icon><Refresh /></el-icon>刷新表结构</el-button></div></template>
      <div class="table-directory" v-loading="tableLoading">
        <div class="directory-toolbar">
          <el-input v-model="keyword" clearable placeholder="筛选数据表" :prefix-icon="Search" />
          <span>{{ filteredTables.length }} 张数据表</span>
        </div>
        <div v-if="filteredTables.length" class="table-grid">
          <button v-for="table in filteredTables" :key="table.name" type="button" class="table-entry" @click="openTable(table.name)">
            <span class="table-entry-icon">▦</span>
            <span class="table-entry-copy"><b>{{ table.name }}</b><small>{{ table.comment || '数据库数据表' }}</small></span>
            <span class="table-entry-arrow">›</span>
          </button>
        </div>
        <el-empty v-else-if="!tableLoading" :image-size="64" description="暂无匹配数据表" />
      </div>
    </el-card>

    <el-dialog v-model="tableDialogVisible" :title="selectedTable" class="database-data-dialog" width="min(1180px, calc(100vw - 40px))" top="7vh" destroy-on-close>
      <template #header><div class="dialog-title"><div><b>{{ selectedTable }}</b><span>{{ total }} 条记录 · {{ columns.length }} 个字段</span></div><el-button :loading="dataLoading" circle title="刷新数据" @click="loadData"><el-icon><Refresh /></el-icon></el-button></div></template>
      <section class="table-content">
          <template v-if="selectedTable">
            <el-table :data="rows" v-loading="dataLoading" border max-height="560" empty-text="该表暂无记录">
              <el-table-column v-for="column in columns" :key="column.name" :prop="column.name" :min-width="columnWidth(column)" show-overflow-tooltip>
                <template #header><div class="column-heading"><span>{{ column.name }}</span><small>{{ column.type }}{{ column.primary_key ? ' · PK' : '' }}</small></div></template>
                <template #default="{ row }"><span class="cell-value">{{ formatValue(row[column.name]) }}</span></template>
              </el-table-column>
            </el-table>
            <div class="pagination"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" @current-change="loadData" @size-change="changePageSize" /></div>
          </template>
      </section>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDatabaseTable, getDatabaseTables } from '@/api/admin'

const tables = ref([])
const selectedTable = ref('')
const columns = ref([])
const rows = ref([])
const total = ref(0)
const tableDialogVisible = ref(false)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const tableLoading = ref(false)
const dataLoading = ref(false)
const filteredTables = computed(() => tables.value.filter(table => `${table.name} ${table.comment || ''}`.toLowerCase().includes(keyword.value.trim().toLowerCase())))

const loadTables = async () => {
  tableLoading.value = true
  try {
    tables.value = await getDatabaseTables()
    if (selectedTable.value && !tables.value.some(table => table.name === selectedTable.value)) selectedTable.value = ''
  } catch { ElMessage.error('加载数据表失败') } finally { tableLoading.value = false }
}

const openTable = async (table) => {
  selectedTable.value = table
  page.value = 1
  tableDialogVisible.value = true
  await loadData()
}

const loadData = async () => {
  if (!selectedTable.value) return
  dataLoading.value = true
  try {
    const data = await getDatabaseTable(selectedTable.value, { page: page.value, page_size: pageSize.value })
    columns.value = data.columns || []
    rows.value = data.rows || []
    total.value = data.total || 0
  } catch { rows.value = []; columns.value = []; total.value = 0 } finally { dataLoading.value = false }
}

const changePageSize = () => { page.value = 1; loadData() }
const columnWidth = (column) => Math.max(130, Math.min(280, (column.name.length + column.type.length) * 9 + 54))
const formatValue = (value) => {
  if (value === null || value === undefined) return 'NULL'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

onMounted(loadTables)
</script>

<style scoped lang="scss">
.database-page { .card-header { display:flex; align-items:center; justify-content:space-between; h3 { margin:0; } p { margin:6px 0 0; color:#909399; font-size:13px; } } }
.table-directory { min-height:560px; padding:20px; border:1px solid #ebeef5; border-radius:8px; background:#fafbfc; }
.directory-toolbar { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px; .el-input { width:300px; } > span { flex:0 0 auto; color:#909399; font-size:13px; } }
.table-grid { display:grid; grid-template-columns:repeat(auto-fill, minmax(260px, 1fr)); gap:12px; }
.table-entry { display:grid; grid-template-columns:38px minmax(0, 1fr) 16px; align-items:center; gap:12px; min-height:80px; padding:12px; border:1px solid #e3e8ef; border-radius:7px; background:#fff; color:#303133; text-align:left; cursor:pointer; transition:border-color .18s ease, box-shadow .18s ease, transform .18s ease; }
.table-entry:hover { border-color:#74aee9; box-shadow:0 8px 18px rgba(63, 120, 180, .12); transform:translateY(-1px); }
.table-entry-icon { display:grid; width:38px; height:38px; place-items:center; border-radius:6px; background:#ecf5ff; color:#409eff; font-size:21px; }
.table-entry-copy { display:grid; min-width:0; gap:5px; } .table-entry-copy b, .table-entry-copy small { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; } .table-entry-copy b { font-family:ui-monospace, SFMono-Regular, Menlo, monospace; font-size:13px; } .table-entry-copy small { color:#909399; font-size:12px; } .table-entry-arrow { color:#a8abb2; font-size:24px; }
.table-content { min-width:0; }
.dialog-title { display:flex; align-items:center; justify-content:space-between; padding-right:26px; } .dialog-title div { display:grid; gap:3px; } .dialog-title b { color:#303133; font-family:ui-monospace, SFMono-Regular, Menlo, monospace; font-size:17px; } .dialog-title span { color:#909399; font-size:12px; }
.column-heading { display:grid; gap:2px; min-width:110px; small { color:#909399; font-size:10px; font-weight:400; } }
.cell-value { font-family:ui-monospace, SFMono-Regular, Menlo, monospace; font-size:12px; white-space:pre-wrap; }
.pagination { display:flex; justify-content:flex-end; margin-top:18px; }
@media (max-width: 700px) { .table-directory { padding:14px; } .directory-toolbar { align-items:stretch; flex-direction:column; gap:10px; .el-input { width:100%; } } .table-grid { grid-template-columns:1fr; } .pagination { justify-content:flex-start; overflow:auto; } }
</style>
