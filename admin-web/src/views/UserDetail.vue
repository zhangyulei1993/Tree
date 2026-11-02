<template>
  <div class="page-stack">
    <PageHeader title="用户详情" description="账号资料与当前家庭绑定。" />
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <section v-if="detail" class="surface detail-card">
      <el-descriptions :column="3" border>
        <el-descriptions-item label="用户 ID">{{ detail.user.id }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.user.phone || '未绑定' }}</el-descriptions-item>
        <el-descriptions-item label="状态"><StatusTag :status="detail.user.status" /></el-descriptions-item>
        <el-descriptions-item label="昵称">{{ detail.user.nickname || '未设置' }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.user.realName || '未设置' }}</el-descriptions-item>
        <el-descriptions-item label="注册端">{{ registerClientText(detail.user.registerClient) }}</el-descriptions-item>
        <el-descriptions-item label="登录方式">{{ detail.user.loginMethod }}</el-descriptions-item>
        <el-descriptions-item label="微信登录">
          <el-tag :type="detail.user.hasWechatLogin ? 'success' : 'info'">
            {{ detail.user.hasWechatLogin ? '已绑定' : '未绑定' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="信任等级">{{ trustTierLabel(detail.user.trustTier) }}</el-descriptions-item>
        <el-descriptions-item label="手机号绑定">
          <el-tag :type="detail.user.phoneVerified ? 'success' : 'info'">
            {{ detail.user.phoneVerified ? '已绑定' : '未绑定' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <h3>当前家庭绑定</h3>
      <el-table :data="detail.families" empty-text="该用户暂未绑定家庭成员">
        <el-table-column prop="familyId" label="家庭 ID" width="100" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="familyName" label="家庭" min-width="180" show-overflow-tooltip />
        <el-table-column prop="memberName" label="成员节点" min-width="160" show-overflow-tooltip />
        <el-table-column label="角色" width="140" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><RoleTag :role="row.familyRole" /></template></el-table-column>
        <el-table-column label="查看" width="180" fixed="right" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center">
          <template #default="{ row }">
            <el-button size="small" @click="router.push(`/admin/families/${row.familyId}`)">家庭详情</el-button>
            <el-button size="small" @click="router.push(`/admin/families/${row.familyId}/tree`)">家庭树</el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getApiErrorMessage } from '@/api/client'
import { getUser } from '@/api/management'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ManagedUserDetail } from '@/types/api'

const route = useRoute()
const router = useRouter()
const detail = ref<ManagedUserDetail | null>(null)
const error = ref('')
function trustTierLabel(tier: string) {
  return tier === 'PHONE_BOUND' ? '已绑定手机号' : '仅微信登录'
}

function registerClientText(client: string) {
  return ({ WECHAT_MINI: '微信小程序', WECHAT_MINI_PROGRAM: '微信小程序', WEB: '网页端', H5: 'H5 页面' } as Record<string, string>)[client] || client || '未记录'
}

async function loadDetail() {
  error.value = ''
  try {
    detail.value = await getUser(String(route.params.userId))
  } catch (e) {
    error.value = getApiErrorMessage(e)
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

</style>
