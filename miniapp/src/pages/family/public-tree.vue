<template>
  <view class="page">
    <view class="card">
      <text class="title">公开家庭树</text>
      <text class="muted">列表树展示，不包含账号、手机号、openid、unionid 或 token。</text>
    </view>
    <view v-for="node in treeNodes" :key="node.memberId" class="card">
      <text class="section-title">{{ node.name }}</text>
      <text class="tag">{{ node.gender }}</text>
      <text class="tag">{{ node.birthText }}</text>
      <text class="muted">{{ node.description }}</text>
      <text class="muted">配偶：{{ names(node.spouseIds) || '无' }}</text>
      <text class="muted">子女：{{ names(node.childrenIds) || '无' }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { treeNodes } from '@/mock/data'

function names(ids: string[]) {
  return ids.map((id) => treeNodes.find((node) => node.memberId === id)?.name).filter(Boolean).join('、')
}
</script>
