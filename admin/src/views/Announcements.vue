<template>
  <div class="announcements-page" v-loading="loading">
    <div class="page-head"><div><h2>全员通知</h2><p>发布后所有用户均可在通知中心查看，不记录用户阅读状态。</p></div><el-button type="primary" @click="openCreate">发布通知</el-button></div>
    <el-table :data="items"><el-table-column prop="title" label="标题" min-width="200"/><el-table-column prop="category" label="分类" width="100"><template #default="{ row }"><el-tag>{{ categoryName(row.category) }}</el-tag></template></el-table-column><el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="row.is_published?'success':'info'">{{ row.is_published?'已发布':'草稿' }}</el-tag></template></el-table-column><el-table-column prop="published_at" label="发布时间" width="180"><template #default="{ row }">{{ row.published_at || '-' }}</template></el-table-column><el-table-column prop="updated_at" label="更新时间" width="180"/><el-table-column label="操作" width="140" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column></el-table>
    <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" layout="total, prev, pager, next" class="pager" @current-change="load"/>
    <el-dialog v-model="visible" :title="form.id?'编辑通知':'发布通知'" width="560px" @closed="reset"><el-form label-position="top"><el-form-item label="标题" required><el-input v-model="form.title" maxlength="120" show-word-limit/></el-form-item><el-form-item label="分类"><el-select v-model="form.category" class="w-full"><el-option label="系统" value="system"/><el-option label="产品" value="product"/><el-option label="活动" value="activity"/></el-select></el-form-item><el-form-item label="内容" required><el-input v-model="form.content" type="textarea" :rows="7" maxlength="5000" show-word-limit/></el-form-item><el-form-item label="立即发布"><el-switch v-model="form.is_published"/><span class="hint">关闭后保存为草稿，用户端不会显示</span></el-form-item></el-form><template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template></el-dialog>
  </div>
</template>
<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const loading=ref(false),saving=ref(false),visible=ref(false),items=ref([])
const pagination=reactive({page:1,pageSize:20,total:0})
const form=reactive({id:null,title:'',content:'',category:'system',is_published:true})
const categoryName=value=>({system:'系统',product:'产品',activity:'活动'}[value]||'系统')
async function load(){loading.value=true;try{const data=await request.get('/announcements',{params:{page:pagination.page,page_size:pagination.pageSize}});items.value=data.list||[];pagination.total=data.total||0}finally{loading.value=false}}
function reset(){Object.assign(form,{id:null,title:'',content:'',category:'system',is_published:true})}
function openCreate(){reset();visible.value=true}
function openEdit(row){Object.assign(form,{id:row.id,title:row.title,content:row.content,category:row.category,is_published:row.is_published});visible.value=true}
async function save(){if(!form.title.trim()||!form.content.trim())return ElMessage.warning('请填写标题和内容');saving.value=true;try{const payload={title:form.title,content:form.content,category:form.category,is_published:form.is_published};if(form.id)await request.put(`/announcements/${form.id}`,payload);else await request.post('/announcements',payload);ElMessage.success('通知已保存');visible.value=false;await load()}finally{saving.value=false}}
async function remove(row){await ElMessageBox.confirm(`确定删除通知“${row.title}”吗？`,'删除通知',{type:'warning'});await request.delete(`/announcements/${row.id}`);ElMessage.success('已删除');await load()}
onMounted(load)
</script>
<style scoped>.announcements-page{padding:20px}.page-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}.page-head h2{margin:0}.page-head p{margin:6px 0 0;color:#909399;font-size:13px}.pager{justify-content:flex-end;margin-top:20px}.hint{margin-left:10px;color:#909399;font-size:12px}</style>
