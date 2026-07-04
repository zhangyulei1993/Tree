<template>
  <div class="page-stack">
    <PageHeader title="用户详情" description="账号资料与当前家庭绑定。" />
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <el-alert v-if="operationError" :title="operationError" type="error" show-icon @close="operationError = ''" />
    <section v-if="detail" class="surface detail-card">
      <el-descriptions :column="3" border>
        <el-descriptions-item label="用户 ID">{{ detail.user.id }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.user.phone || '未绑定' }}</el-descriptions-item>
        <el-descriptions-item label="状态"><StatusTag :status="detail.user.status" /></el-descriptions-item>
        <el-descriptions-item label="昵称">{{ detail.user.nickname || '未设置' }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.user.realName || '未设置' }}</el-descriptions-item>
        <el-descriptions-item label="注册端">{{ detail.user.registerClient }}</el-descriptions-item>
        <el-descriptions-item label="登录方式">{{ detail.user.loginMethod }}</el-descriptions-item>
        <el-descriptions-item label="微信登录">
          <el-tag :type="detail.user.hasWechatLogin ? 'success' : 'info'">
            {{ detail.user.hasWechatLogin ? '已绑定' : '未绑定' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="信任等级">{{ trustTierLabel(detail.user.trustTier) }}</el-descriptions-item>
        <el-descriptions-item label="手机号登录">
          <el-tag :type="detail.user.phoneLoginEnabled ? 'success' : 'info'">
            {{ detail.user.phoneLoginEnabled ? '已开启' : '未开启' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="canUnbindPhoneLogin" class="action-row">
        <el-button type="danger" :loading="unbinding" @click="confirmUnbind">解除手机号登录</el-button>
        <span class="hint">解除后用户仅保留微信登录，需重新设置手机号与密码方可恢复。</span>
      </div>

      <h3>当前家庭绑定</h3>
      <el-table :data="detail.families">
        <el-table-column prop="familyId" label="家庭 ID" />
        <el-table-column prop="familyName" label="家庭" />
        <el-table-column prop="memberName" label="成员节点" />
        <el-table-column prop="familyRole" label="角色" />
      </el-table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import { getApiErrorMessage } from '@/api/client'
import { getUser, unbindPhoneLogin } from '@/api/management'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useAuthStore } from '@/stores/auth'
import type { ManagedUserDetail } from '@/types/api'

const route = useRoute()
const auth = useAuthStore()
const detail = ref<ManagedUserDetail | null>(null)
const error = ref('')
const operationError = ref('')
const unbinding = ref(false)

const canUnbindPhoneLogin = computed(
  () =>
    auth.hasRole(['ROOT_ADMIN', 'SUPER_ADMIN']) &&
    Boolean(detail.value?.user.canUnbindPhoneLogin)
)

function trustTierLabel(tier: string) {
  return tier === 'PHONE_BOUND' ? '备用登录已开启' : '仅微信'
}

async function loadDetail() {
  error.value = ''
  try {
    detail.value = await getUser(String(route.params.userId))
  } catch (e) {
    error.value = getApiErrorMessage(e)
  }
}

async function confirmUnbind() {
  if (!detail.value) return
  try {
    await ElMessageBox.confirm(
      `确认解除用户 ${detail.value.user.id} 的手机号登录？解除后该用户将无法使用手机号密码登录，信任等级将降为「仅微信」。`,
      '解除手机号登录',
      { type: 'warning', confirmButtonText: '确认解除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }

  unbinding.value = true
  operationError.value = ''
  try {
    await unbindPhoneLogin(detail.value.user.id, '管理员手动解除手机号登录')
    ElMessage.success('已解除手机号登录')
    await loadDetail()
  } catch (e) {
    operationError.value = getApiErrorMessage(e)
  } finally {
    unbinding.value = false
  }
}

onMounted(loadDetail)
</script>

<style scoped>
.detail-card {
  padding: 16px;
}

h3 {
  margin-top: 22px;
}

.action-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
}

.hint {
  color: #7a8797;
  font-size: 13px;
}
</style>
