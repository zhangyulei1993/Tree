<template>
  <view class="page">
    <view class="card">
      <text class="title">我的家庭</text>
      <text class="muted">展示我加入或创建的家庭 mock 数据。</text>
    </view>
    <view v-for="family in myFamilies" :key="family.id" class="card" @click="go('/pages/family/detail')">
      <text class="section-title">{{ family.name }}</text>
      <text class="tag">{{ family.role }}</text>
      <text class="muted">成员节点：{{ family.memberName }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { myFamilies } from '@/mock/data'

function go(url: string) {
  uni.navigateTo({ url })
}
</script>
