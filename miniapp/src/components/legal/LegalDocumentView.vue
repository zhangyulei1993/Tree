<template>
  <view class="tree-page legal-page">
    <MiniBackHome />
    <view class="legal-lead card">
      <text class="title">{{ title }}</text>
      <text class="muted">{{ metaLine }}</text>
    </view>
    <view v-for="section in sections" :key="section.title" class="card legal-section">
      <text class="section-title">{{ section.title }}</text>
      <text
        v-for="(paragraph, index) in section.paragraphs"
        :key="`${section.title}-${index}`"
        class="body"
      >
        {{ paragraph }}
      </text>
    </view>
    <view v-if="appendSections.length" class="card legal-section">
      <text class="section-title">附录：信息收集明细</text>
      <view v-for="item in appendSections" :key="item.category" class="legal-append-item">
        <text class="append-title">{{ item.category }}</text>
        <text class="body">信息类型：{{ item.fields }}</text>
        <text class="body">处理目的：{{ item.purpose }}</text>
        <text class="body">处理方式：{{ item.method }}</text>
        <text class="body">必要性：{{ item.necessary }}</text>
        <text class="body">保存期限：{{ item.retention }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import type { CollectedInfoItem } from '@/features/legal/privacyCatalog'
import type { LegalSection } from '@/features/legal/legalContent'

defineProps<{
  title: string
  metaLine: string
  sections: LegalSection[]
  appendSections?: CollectedInfoItem[]
}>()
</script>

<style scoped>
.legal-lead {
  margin-bottom: 20rpx;
}

.legal-section {
  margin-bottom: 20rpx;
}

.body {
  display: block;
  margin-top: 12rpx;
  color: var(--tree-text-secondary);
  font-size: 27rpx;
  line-height: 1.9;
  white-space: pre-wrap;
}

.legal-append-item + .legal-append-item {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 1rpx solid var(--tree-border-subtle);
}

.append-title {
  display: block;
  margin-bottom: 8rpx;
  color: var(--tree-text);
  font-size: 28rpx;
  font-weight: 600;
}
</style>
