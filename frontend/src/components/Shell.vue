<template>
  <el-container class="h-screen">
    <el-aside width="220px" class="border-r">
      <div class="p-4 text-center text-lg font-bold">工时系统</div>
      <el-menu router :default-active="$route.path" class="border-0">
        <el-menu-item index="/timesheet">录入工时</el-menu-item>
        <el-menu-item index="/ai/report">AI生成日报</el-menu-item>
        <el-menu-item index="/work-plan">工作计划</el-menu-item>
        <el-menu-item index="/work-schedule">工作排期</el-menu-item>
        <el-menu-item index="/account/security">账号安全</el-menu-item>
        <template v-if="role === 'admin'">
          <el-menu-item index="/admin/projects">项目管理</el-menu-item>
          <el-menu-item index="/admin/users">用户管理</el-menu-item>
          <el-menu-item index="/admin/report">工时报表</el-menu-item>
          <el-menu-item index="/admin/timesheet-backfill">工时补入</el-menu-item>
          <el-menu-item index="/admin/timesheet-backfill-history">补入历史</el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="flex items-center justify-between px-6">
        <div class="text-base text-gray-600">欢迎使用</div>
        <div class="flex items-center gap-3">
          <el-tag type="success" v-if="role === 'admin'">管理员</el-tag>
          <el-button size="small" @click="logout">退出</el-button>
        </div>
      </el-header>
      <el-main class="bg-gray-50">
        <div :class="fullWidth ? 'p-4' : 'max-w-6xl mx-auto p-4'">
          <slot />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '../store/auth'

defineProps<{ fullWidth?: boolean }>()

const auth = useAuthStore()
const { role } = storeToRefs(auth)
const logout = () => auth.logout()
</script>

<style scoped>
.h-screen {
  height: 100vh;
}

.border-r {
  border-right: 1px solid #eee;
}

.bg-gray-50 {
  background: #fafafa;
}

.text-gray-600 {
  color: #606266;
}

.max-w-6xl {
  max-width: 1152px;
}

.mx-auto {
  margin-left: auto;
  margin-right: auto;
}

.p-4 {
  padding: 1rem;
}

.px-6 {
  padding-left: 1.5rem;
  padding-right: 1.5rem;
}

.flex {
  display: flex;
}

.items-center {
  align-items: center;
}

.justify-between {
  justify-content: space-between;
}

.gap-3 {
  gap: .75rem;
}

.text-lg {
  font-size: 1.125rem;
}

.font-bold {
  font-weight: 700;
}

.border-0 {
  border: 0;
}
</style>