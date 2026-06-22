<template>
  <view class="tree-page">
    <MiniSectionHeader title="家庭搜索" subtitle="查找已公开的家庭主页。" />

    <MiniCard>
      <template v-if="isRealApiMode">
        <MiniNotice tone="warm" title="搜索功能即将开放">
          公开家庭搜索暂未开放，请通过邀请链接或公开家庭链接访问。
        </MiniNotice>
      </template>
      <template v-else>
        <text class="tree-field-label">搜索关键词</text>
        <input v-model="keyword" class="tree-input" placeholder="家族名称 / 姓氏 / 地区" />
      </template>
    </MiniCard>

    <template v-if="!isRealApiMode">
      <MiniCard v-if="filtered.length === 0">
        <MiniEmptyState
          symbol="寻"
          :title="keyword ? '未找到匹配家庭' : '暂无公开家庭'"
          :description="keyword ? '请尝试其他关键词，或通过邀请链接访问家庭。' : '当前没有可浏览的公开家庭。'"
        />
      </MiniCard>

      <FamilyMiniCard
        v-for="family in filtered"
        :key="family.id"
        :name="family.name"
        :surname="family.surname"
        :region="family.regionText"
        :desc="family.description"
        @click="openFamily(family.id)"
      />
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { isRealApiMode } from '@/api/client'
import MiniCard from '@/components/base/MiniCard.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import MiniSectionHeader from '@/components/base/MiniSectionHeader.vue'
import FamilyMiniCard from '@/components/family/FamilyMiniCard.vue'
import { families } from '@/mock/data'

const keyword = ref('')
const filtered = computed(() => families.filter((item) => !keyword.value || JSON.stringify(item).includes(keyword.value)))

function openFamily(familyId: string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(familyId)}`
  })
}
</script>
