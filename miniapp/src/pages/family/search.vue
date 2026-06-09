<template>
  <view class="page">
    <view class="card">
      <text class="title">家庭搜索</text>
      <input v-model="keyword" class="input" placeholder="家族名称 / 姓氏 / 地区" />
    </view>
    <view v-for="family in filtered" :key="family.id" class="card" @click="go('/pages/family/public-profile')">
      <text class="section-title">{{ family.name }}</text>
      <text class="tag">{{ family.surname }}</text>
      <text class="tag">{{ family.regionText }}</text>
      <text class="muted">{{ family.description }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

import { families } from '@/mock/data'

const keyword = ref('')
const filtered = computed(() => families.filter((item) => !keyword.value || JSON.stringify(item).includes(keyword.value)))

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
