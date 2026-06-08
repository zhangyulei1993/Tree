<template>
  <section class="surface member-detail">
    <template v-if="member">
      <div class="title-row">
        <h3>{{ member.name }}</h3>
        <RoleTag :role="member.role" />
      </div>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="成员 ID">{{ member.id }}</el-descriptions-item>
        <el-descriptions-item label="性别">{{ member.gender }}</el-descriptions-item>
        <el-descriptions-item label="出生年份">{{ member.birthYear }}</el-descriptions-item>
        <el-descriptions-item label="绑定状态">{{ member.bindingStatus }}</el-descriptions-item>
        <el-descriptions-item label="成员状态">
          <StatusTag :status="member.status" />
        </el-descriptions-item>
        <el-descriptions-item label="说明">{{ member.description }}</el-descriptions-item>
      </el-descriptions>
      <div class="table-actions">
        <el-button>编辑</el-button>
        <el-button>站内邀请</el-button>
        <el-button>添加父亲</el-button>
        <el-button>添加母亲</el-button>
        <el-button>添加子女</el-button>
        <el-button>添加配偶</el-button>
        <el-button type="danger">删除</el-button>
      </div>
    </template>
    <el-empty v-else description="请选择左侧成员" />
  </section>
</template>

<script setup lang="ts">
import RoleTag from '@/components/RoleTag.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { MemberNode } from '@/mock/data'

defineProps<{ member?: MemberNode }>()
</script>

<style scoped>
.member-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h3 {
  margin: 0;
}
</style>
