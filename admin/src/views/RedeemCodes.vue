<template>
  <div class="redeem-codes">
    <div class="header"><div><h2>兑换码管理</h2><p>创建灵感值兑换码，并查看有效期与用户兑换记录。</p></div><el-button type="primary" @click="openCreate">创建兑换码</el-button></div>
    <el-form :inline="true" class="filters"><el-form-item label="搜索"><el-input v-model.trim="filters.keyword" clearable placeholder="兑换码或批次名称" @keyup.enter="search" /></el-form-item><el-form-item label="状态"><el-select v-model="filters.status" clearable placeholder="全部"><el-option label="未兑换" value="unused"/><el-option label="已兑换" value="used"/><el-option label="已过期" value="expired"/></el-select></el-form-item><el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button></el-form-item></el-form>
    <el-table :data="items" v-loading="loading" empty-text="暂无兑换码"><el-table-column prop="code" label="兑换码" min-width="190"><template #default="{row}"><code>{{ row.code }}</code></template></el-table-column><el-table-column prop="amount" label="灵感值" width="100"/><el-table-column prop="batch_name" label="批次" min-width="140"/><el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="status(row).type">{{ status(row).text }}</el-tag></template></el-table-column><el-table-column label="有效期" width="180"><template #default="{row}">{{ row.expires_at || '永久有效' }}</template></el-table-column><el-table-column label="兑换邮箱" min-width="200"><template #default="{row}">{{ row.used_email || '-' }}</template></el-table-column><el-table-column label="兑换时间" width="180"><template #default="{row}">{{ row.used_at || '-' }}</template></el-table-column><el-table-column prop="created_at" label="创建时间" width="180"/></el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total, prev, pager, next" @current-change="fetchItems"/>
    <el-dialog v-model="createVisible" title="创建兑换码" width="480px"><el-form :model="form" label-width="90px"><el-form-item label="批次名称"><el-input v-model.trim="form.batch_name" maxlength="100" placeholder="留空自动生成"/><div class="form-tip">默认生成 AGI-XXXXXX，可直接修改。</div></el-form-item><el-form-item label="单码灵感值" required><el-input-number v-model="form.amount" :min="1" :max="1000000"/></el-form-item><el-form-item label="生成数量" required><el-input-number v-model="form.quantity" :min="1" :max="1000"/></el-form-item><el-form-item label="有效期"><el-date-picker v-model="form.expires_at" type="datetime" clearable value-format="YYYY-MM-DD HH:mm:ss" placeholder="留空表示永久有效"/></el-form-item></el-form><template #footer><el-button @click="createVisible=false">取消</el-button><el-button type="primary" :loading="creating" @click="create">创建</el-button></template></el-dialog>
    <el-dialog v-model="resultVisible" title="本批次兑换码" width="560px" @closed="generatedCodes=[]"><div class="result-header"><span>共 {{ generatedCodes.length }} 个，可逐条复制或全部复制。</span><el-button type="primary" plain @click="copyAll">复制全部</el-button></div><div class="generated-list"><div v-for="item in generatedCodes" :key="item.id" class="generated-row"><code>{{ item.code }}</code><el-button link type="primary" @click="copyCode(item.code)">复制</el-button></div></div><template #footer><el-button type="primary" @click="resultVisible=false">完成</el-button></template></el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'
const loading = ref(false), creating = ref(false), createVisible = ref(false), resultVisible = ref(false), items = ref([]), generatedCodes = ref([])
const filters = reactive({ keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const form = reactive({ batch_name: '', amount: 10, quantity: 1, expires_at: '' })
const fetchItems = async () => { loading.value = true; try { const data = await request.get('/redeem-codes', { params: { ...filters, page: pagination.page, page_size: pagination.pageSize } }); items.value = data.list || []; pagination.total = data.total || 0 } finally { loading.value = false } }
const search = () => { pagination.page = 1; fetchItems() }
const reset = () => { Object.assign(filters, { keyword: '', status: '' }); search() }
const randomBatchName = () => `AGI-${Math.random().toString(16).slice(2, 8).toUpperCase().padEnd(6, '0')}`
const openCreate = () => { Object.assign(form, { batch_name: randomBatchName(), amount: 10, quantity: 1, expires_at: '' }); createVisible.value = true }
const copyCode = async (code) => { try { await navigator.clipboard.writeText(code); ElMessage.success('兑换码已复制') } catch { ElMessage.error('复制失败，请手动复制') } }
const copyAll = async () => { try { await navigator.clipboard.writeText(generatedCodes.value.map(item => item.code).join('\n')); ElMessage.success('全部兑换码已复制') } catch { ElMessage.error('复制失败，请手动复制') } }
const create = async () => { creating.value = true; try { const created = await request.post('/redeem-codes', form); generatedCodes.value = created || []; ElMessage.success('兑换码创建成功'); createVisible.value = false; resultVisible.value = true; await fetchItems() } finally { creating.value = false } }
const status = (row) => { if (row.used_by) return { text: '已兑换', type: 'success' }; if (row.expires_at && new Date(row.expires_at.replace(/-/g, '/')) < new Date()) return { text: '已过期', type: 'info' }; return { text: '未兑换', type: 'warning' } }
onMounted(fetchItems)
</script>

<style scoped>
.redeem-codes { padding: 20px; }.header { display:flex; justify-content:space-between; align-items:center; margin-bottom:20px; }.header h2 { margin:0; }.header p { margin:6px 0 0; color:#909399; font-size:13px; }.filters { margin-bottom:18px; }.el-pagination { margin-top:20px; justify-content:flex-end; } code { color:#409eff; }.form-tip { margin-top:5px; color:#909399; font-size:12px; line-height:1.4; }.result-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:12px; color:#606266; font-size:13px; }.generated-list { max-height:380px; overflow:auto; border:1px solid #ebeef5; border-radius:4px; }.generated-row { display:flex; align-items:center; justify-content:space-between; min-height:42px; padding:0 12px; border-bottom:1px solid #ebeef5; }.generated-row:last-child { border-bottom:0; }.generated-row code { user-select:all; }
</style>
