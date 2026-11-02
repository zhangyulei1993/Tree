<template>
  <div class="page-stack">
    <PageHeader title="家庭成员索引" description="用于快速检索与定位；成员资料、亲属关系和后续变更统一回到家庭树卡片处理。">
      <el-button @click="router.push(`/admin/families/${familyId}`)">返回家庭详情</el-button>
      <el-button @click="router.push(`/admin/families/${familyId}/tree`)">查看家庭树</el-button>
    </PageHeader>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <DataTable>
      <el-table v-loading="loading" :data="items" :row-class-name="memberRowClassName" stripe empty-text="该家庭暂未创建成员节点">
        <el-table-column prop="id" label="成员 ID" width="90" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center" />
        <el-table-column prop="displayName" label="姓名" min-width="180" show-overflow-tooltip />
        <el-table-column prop="gender" label="性别" width="80" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }">{{ row.gender === 'MALE' ? '男' : row.gender === 'FEMALE' ? '女' : '未知' }}</template></el-table-column>
        <el-table-column prop="familyRole" label="角色" width="140" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><RoleTag :role="row.familyRole || 'MEMBER'" /></template></el-table-column>
        <el-table-column label="绑定策略" width="130" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }">{{ bindingPolicyText(row.userBindingPolicy) }}</template></el-table-column>
        <el-table-column label="账号绑定" width="100" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }">{{ row.boundUserId ? '已绑定' : '未绑定' }}</template></el-table-column>
        <el-table-column label="状态" width="96" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><StatusTag :status="row.status" /></template></el-table-column>
        <el-table-column label="操作" width="132" fixed="right" align="center" header-align="center" class-name="table-column--center" label-class-name="table-column--center"><template #default="{ row }"><el-button size="small" type="primary" plain @click="locateInTree(row.id)">定位到树</el-button></template></el-table-column>
      </el-table>
    </DataTable>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getApiErrorMessage } from '@/api/client'
import { listFamilyMembers } from '@/api/management'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import RoleTag from '@/components/RoleTag.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ManagedFamilyMember } from '@/types/api'

const route = useRoute()
const router = useRouter()
const familyId = String(route.params.familyId)
const highlightedMemberId = Number(route.query.memberId)
const items = ref<ManagedFamilyMember[]>([])
const loading = ref(true)
const error = ref('')

function bindingPolicyText(policy: string) {
  return ({ REQUIRED: '需要绑定', OPTIONAL: '可选绑定', NOT_REQUIRED: '无需绑定' } as Record<string, string>)[policy] || policy || '未设置'
}

function memberRowClassName({ row }: { row: ManagedFamilyMember }) {
  return highlightedMemberId === row.id ? 'is-located-member' : ''
}

function locateInTree(memberId: number) {
  router.push({ path: `/admin/families/${familyId}/tree`, query: { memberId: String(memberId) } })
}

onMounted(async () => {
  try {
    items.value = await listFamilyMembers(familyId)
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
:deep(.is-located-member td) {
  background: rgba(182, 71, 62, .08) !important;
}
</style>
