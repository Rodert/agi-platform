<template>
  <div class="permissions">
    <h2>权限管理</h2>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="角色管理" name="roles">
        <el-button type="primary" @click="addRole" style="margin-bottom: 20px">
          添加角色
        </el-button>

        <el-table :data="roles" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="角色名称" width="150" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="user_count" label="用户数" width="100" />
          <el-table-column prop="created_at" label="创建时间" width="180" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="editRole(row)">编辑</el-button>
              <el-button size="small" @click="setPermissions(row)">权限</el-button>
              <el-button size="small" type="danger" @click="deleteRole(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="权限列表" name="permissions">
        <el-table :data="permissions" style="width: 100%" row-key="id" default-expand-all>
          <el-table-column prop="name" label="权限名称" width="200" />
          <el-table-column prop="code" label="权限代码" width="200" />
          <el-table-column prop="description" label="描述" />
          <el-table-column prop="module" label="所属模块" width="120" />
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="管理员列表" name="admins">
        <el-button type="primary" @click="addAdmin" style="margin-bottom: 20px">
          添加管理员
        </el-button>

        <el-table :data="admins" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="name" label="姓名" width="120" />
          <el-table-column prop="role" label="角色" width="120">
            <template #default="{ row }">
              <el-tag>{{ row.role }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_active ? 'success' : 'danger'">
                {{ row.is_active ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="last_login_at" label="最后登录" width="180" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="editAdmin(row)">编辑</el-button>
              <el-button
                size="small"
                :type="row.is_active ? 'danger' : 'success'"
                @click="toggleAdmin(row)"
              >
                {{ row.is_active ? '禁用' : '启用' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('roles')

const roles = ref([
  {
    id: 1,
    name: '超级管理员',
    description: '拥有所有权限',
    user_count: 2,
    created_at: '2026-01-01 00:00:00'
  },
  {
    id: 2,
    name: '运营人员',
    description: '负责内容审核和用户管理',
    user_count: 5,
    created_at: '2026-01-01 00:00:00'
  }
])

const permissions = ref([
  {
    id: 1,
    name: '用户管理',
    code: 'user.manage',
    description: '用户管理权限',
    module: '用户模块',
    children: [
      { id: 11, name: '查看用户', code: 'user.view', description: '查看用户列表', module: '用户模块' },
      { id: 12, name: '编辑用户', code: 'user.edit', description: '编辑用户信息', module: '用户模块' },
      { id: 13, name: '删除用户', code: 'user.delete', description: '删除用户', module: '用户模块' }
    ]
  },
  {
    id: 2,
    name: '作品管理',
    code: 'work.manage',
    description: '作品管理权限',
    module: '作品模块',
    children: [
      { id: 21, name: '查看作品', code: 'work.view', description: '查看作品列表', module: '作品模块' },
      { id: 22, name: '审核作品', code: 'work.audit', description: '审核作品', module: '作品模块' },
      { id: 23, name: '删除作品', code: 'work.delete', description: '删除作品', module: '作品模块' }
    ]
  }
])

const admins = ref([
  {
    id: 1,
    username: 'admin',
    name: '超级管理员',
    role: '超级管理员',
    is_active: true,
    last_login_at: '2026-07-12 20:00:00'
  }
])

const addRole = () => {
  ElMessage.info('添加角色功能待实现')
}

const editRole = (row) => {
  ElMessage.info(`编辑角色: ${row.name}`)
}

const setPermissions = (row) => {
  ElMessage.info(`设置角色权限: ${row.name}`)
}

const deleteRole = (row) => {
  ElMessage.warning(`删除角色: ${row.name}`)
}

const addAdmin = () => {
  ElMessage.info('添加管理员功能待实现')
}

const editAdmin = (row) => {
  ElMessage.info(`编辑管理员: ${row.username}`)
}

const toggleAdmin = (row) => {
  ElMessage.success(`${row.is_active ? '禁用' : '启用'}管理员: ${row.username}`)
}
</script>

<style scoped>
.permissions {
  padding: 20px;
}
</style>
