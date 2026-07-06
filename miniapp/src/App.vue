<template>
  <PrivacyConsentModal />
</template>

<script setup lang="ts">
import { onLaunch, onShow } from '@dcloudio/uni-app'

import PrivacyConsentModal from '@/components/legal/PrivacyConsentModal.vue'
import { promptPrivacyConsentIfNeeded } from '@/features/legal/privacyConsent'

onLaunch(() => {
  promptPrivacyConsentIfNeeded()
})

onShow(() => {
  promptPrivacyConsentIfNeeded()
})
</script>

<style>
@import './styles/theme.css';
@import './styles/components.css';

page {
  background: linear-gradient(180deg, var(--tree-bg-top) 0%, var(--tree-bg-mid) 45%, var(--tree-bg-bottom) 100%);
  color: var(--tree-text);
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.page,
.tree-page {
  position: relative;
  min-height: 100vh;
  padding: var(--tree-space-page);
  padding-bottom: calc(132rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
  background: linear-gradient(180deg, var(--tree-bg-top) 0%, var(--tree-bg-mid) 50%, var(--tree-bg-bottom) 100%);
}

.auth-page {
  overflow-x: hidden;
  padding-top: 36rpx;
  background: #f3f6f3 !important;
}

.auth-page::before,
.auth-page::after {
  display: none !important;
  content: none !important;
  width: 0 !important;
  height: 0 !important;
  opacity: 0 !important;
}

.auth-page .mini-back-home {
  display: none !important;
}

.auth-page .mini-card.soft {
  background: rgba(250, 252, 251, 0.98) !important;
  box-shadow: none !important;
}

.card,
.tree-card {
  margin-bottom: 20rpx;
  border: 1rpx solid var(--tree-border-subtle);
  border-radius: var(--tree-radius-lg);
  background: var(--tree-surface);
  padding: var(--tree-space-card);
  box-shadow: var(--tree-shadow-sm);
}

.title,
.tree-page-title {
  display: block;
  margin-bottom: 10rpx;
  color: var(--tree-text-primary);
  font-size: 36rpx;
  font-weight: 700;
  line-height: 1.35;
}

.section-title,
.tree-section-title {
  display: block;
  margin-bottom: 14rpx;
  color: var(--tree-text);
  font-size: 30rpx;
  font-weight: 600;
}

.muted,
.tree-muted {
  color: var(--tree-text-secondary);
  font-size: 26rpx;
  line-height: 1.7;
}

.tree-weak {
  color: var(--tree-text-weak);
  font-size: 24rpx;
  line-height: 1.6;
}

.tree-field-label {
  display: block;
  margin-top: 18rpx;
  margin-bottom: 8rpx;
  color: var(--tree-text-secondary);
  font-size: 24rpx;
}

.tree-field-label:first-of-type {
  margin-top: 8rpx;
}

.tree-field-error {
  display: block;
  margin-top: 8rpx;
  color: #b4533a;
  font-size: 24rpx;
}

.tree-field-success {
  display: block;
  margin-top: 8rpx;
  color: var(--tree-green);
  font-size: 24rpx;
}

.tag {
  display: inline-flex;
  margin-right: 10rpx;
  margin-bottom: 10rpx;
  border-radius: 999rpx;
  background: var(--tree-green-light);
  color: var(--tree-green);
  padding: 8rpx 18rpx;
  font-size: 24rpx;
}

.button,
.tree-primary-button {
  margin: 16rpx 0 0;
  border-radius: var(--tree-radius-md);
  background: linear-gradient(135deg, var(--tree-primary) 0%, var(--tree-green) 100%);
  color: #fff;
  font-weight: 600;
}

.button.secondary,
.tree-secondary-button {
  background: #fff;
  color: var(--tree-text-primary);
  border: 1rpx solid var(--tree-border-warm);
}

.button.danger,
.tree-danger-button {
  background: #b4533a;
  color: #fff;
}

.button[disabled],
.tree-primary-button[disabled],
.tree-secondary-button[disabled] {
  opacity: 0.55;
}

.input,
.textarea,
.tree-input,
.tree-textarea {
  width: 100%;
  box-sizing: border-box;
  margin-top: 12rpx;
  border: 1rpx solid var(--tree-border-warm);
  border-radius: var(--tree-radius-md);
  background: #fff;
  font-size: 28rpx;
}

.input,
.tree-input {
  height: 80rpx;
  min-height: 80rpx;
  line-height: 80rpx;
  padding: 0 22rpx;
}

.textarea,
.tree-textarea {
  padding: 20rpx 22rpx;
  min-height: 180rpx;
}

.tree-notice {
  border-radius: var(--tree-radius-md);
  padding: 18rpx 22rpx;
  background: var(--tree-green-light);
  border: 1rpx solid var(--tree-border-green);
  color: var(--tree-text-secondary);
  font-size: 24rpx;
  line-height: 1.7;
}

.tree-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14rpx;
}

.tree-section {
  margin-bottom: 28rpx;
}

.tree-info-row {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  padding: 16rpx 0;
  border-top: 1rpx solid var(--tree-border-subtle);
  font-size: 26rpx;
}

.tree-info-row:first-child {
  border-top: none;
  padding-top: 0;
}

.tree-info-label {
  flex-shrink: 0;
  color: var(--tree-text-secondary);
}

.tree-info-value {
  color: var(--tree-text);
  text-align: right;
  max-width: 62%;
  line-height: 1.5;
}

@media screen and (min-width: 960px) {
  .page,
  .tree-page {
    max-width: 960px;
    margin: 0 auto;
    padding-inline: 24px;
  }
}
</style>
