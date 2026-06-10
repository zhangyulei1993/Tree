<template>
  <view class="page">
    <view class="card">
      <text class="title">家庭搜索</text>
      <template v-if="isRealApiMode">
        <text class="muted">公开家庭搜索暂未开放，请通过邀请链接或公开家庭链接访问。</text>
      </template>
      <template v-else>
        <input v-model="keyword" class="input" placeholder="家族名称 / 姓氏 / 地区" />
      </template>
    </view>
    <template v-if="!isRealApiMode">
      <view v-for="family in filtered" :key="family.id" class="card" @click="openFamily(family.id)">
        <text class="section-title">{{ family.name }}</text>
        <text class="tag">{{ family.surname }}</text>
        <text class="tag">{{ family.regionText }}</text>
        <text class="muted">{{ family.description }}</text>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { isRealApiMode } from '@/api/client'
import { families } from '@/mock/data'

const keyword = ref('')
const filtered = computed(() => families.filter((item) => !keyword.value || JSON.stringify(item).includes(keyword.value)))

function openFamily(familyId: string) {
  uni.navigateTo({
    url: `/pages/family/public-profile?familyId=${encodeURIComponent(familyId)}`
  })
}
</script>
