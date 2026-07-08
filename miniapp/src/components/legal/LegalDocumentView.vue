<template>
  <view class="archive-page legal-page">
    <MiniBackHome />

    <view class="legal-head archive-page-head">
      <view>
        <text class="archive-kicker">{{ kicker }}</text>
        <text class="archive-title">{{ title }}</text>
        <text class="legal-product">{{ productName }}</text>
        <text class="legal-updated">更新日期 {{ updatedDate }}</text>
        <text class="legal-meta-line">{{ metaLine }}</text>
      </view>
      <view class="archive-seal">{{ sealLetter }}</view>
    </view>

    <view class="legal-body">
      <view v-for="section in sections" :key="section.title" class="legal-section">
        <text class="legal-section-title">{{ section.title }}</text>
        <text
          v-for="(paragraph, index) in section.paragraphs"
          :key="`${section.title}-${index}`"
          class="legal-paragraph"
        >
          {{ paragraph }}
        </text>
      </view>

      <view v-if="appendSections.length" class="legal-section legal-appendix">
        <text class="legal-section-title">附录：信息收集明细</text>
        <view v-for="item in appendSections" :key="item.category" class="legal-append-item">
          <text class="legal-append-title">{{ item.category }}</text>
          <text class="legal-paragraph">信息类型：{{ item.fields }}</text>
          <text class="legal-paragraph">处理目的：{{ item.purpose }}</text>
          <text class="legal-paragraph">处理方式：{{ item.method }}</text>
          <text class="legal-paragraph">必要性：{{ item.necessary }}</text>
          <text class="legal-paragraph">保存期限：{{ item.retention }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import type { LegalSection } from '@/features/legal/legalContent'
import type { CollectedInfoItem } from '@/features/legal/privacyCatalog'

withDefaults(
  defineProps<{
    title: string
    kicker: string
    sealLetter: string
    productName: string
    updatedDate: string
    metaLine: string
    sections: LegalSection[]
    appendSections?: CollectedInfoItem[]
  }>(),
  {
    appendSections: () => []
  }
)
</script>

<style scoped>
.legal-page {
  padding-top: 28rpx;
}

.legal-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.legal-product {
  display: block;
  margin-top: 12rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 700;
  line-height: 1.4;
}

.legal-updated {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-cinnabar);
  font-size: 22rpx;
  line-height: 1.45;
}

.legal-meta-line {
  display: block;
  margin-top: 8rpx;
  color: var(--archive-ink-soft);
  font-size: 21rpx;
  line-height: 1.5;
}

.legal-body {
  margin-top: 8rpx;
  border-top: 1rpx solid var(--archive-line-strong);
  border-bottom: 1rpx solid var(--archive-line);
  background: rgba(255, 252, 245, 0.36);
}

.legal-section {
  padding: 28rpx 0;
  border-bottom: 1rpx solid var(--archive-line);
}

.legal-section:last-child {
  border-bottom: 0;
}

.legal-section-title {
  display: block;
  margin-bottom: 4rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 30rpx;
  font-weight: 700;
  line-height: 1.45;
}

.legal-paragraph {
  display: block;
  margin-top: 14rpx;
  color: var(--archive-ink-soft);
  font-size: 26rpx;
  line-height: 2;
  white-space: pre-wrap;
}

.legal-appendix {
  background: rgba(255, 249, 236, 0.28);
}

.legal-append-item {
  margin-top: 22rpx;
  padding-top: 22rpx;
  border-top: 1rpx solid var(--archive-line);
}

.legal-append-item:first-of-type {
  margin-top: 12rpx;
  padding-top: 0;
  border-top: 0;
}

.legal-append-title {
  display: block;
  margin-bottom: 4rpx;
  color: var(--archive-ink);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 28rpx;
  font-weight: 650;
  line-height: 1.4;
}
</style>
