<template>
  <view class="family-page-skeleton">
    <template v-if="variant === 'detail'">
      <view class="sk-card sk-hero">
        <view class="sk-hero-head">
          <view class="sk-circle" />
          <view class="sk-col">
            <view class="sk-line sk-w30" />
            <view class="sk-line sk-w70" />
            <view class="sk-tags">
              <view class="sk-pill" />
              <view class="sk-pill" />
            </view>
          </view>
        </view>
        <view class="sk-line sk-w90" />
        <view class="sk-line sk-w50" />
      </view>
      <view class="sk-card sk-soft">
        <view class="sk-info-row">
          <view class="sk-line sk-w35" />
          <view class="sk-line sk-w25" />
        </view>
        <view class="sk-info-row">
          <view class="sk-line sk-w35" />
          <view class="sk-line sk-w25" />
        </view>
      </view>
      <view class="sk-card">
        <view v-for="i in 4" :key="i" class="sk-directory-row">
          <view class="sk-icon" />
          <view class="sk-col">
            <view class="sk-line sk-w45" />
            <view class="sk-line sk-w65" />
          </view>
        </view>
      </view>
    </template>

    <template v-else-if="variant === 'profile'">
      <view class="sk-card sk-context">
        <view class="sk-line sk-w25" />
        <view class="sk-line sk-w55" />
        <view class="sk-line sk-w40" />
      </view>
      <view class="sk-card">
        <view class="sk-space-head">
          <view class="sk-col">
            <view class="sk-line sk-w40" />
            <view class="sk-line sk-w60" />
          </view>
          <view class="sk-pill sk-pill--action" />
        </view>
        <view v-for="i in 7" :key="i" class="sk-info-row">
          <view class="sk-line sk-w30" />
          <view class="sk-line sk-w45" />
        </view>
      </view>
    </template>

    <template v-else>
      <view class="sk-card sk-context">
        <view class="sk-line sk-w25" />
        <view class="sk-line sk-w55" />
        <view class="sk-line sk-w40" />
      </view>
      <view class="sk-card">
        <view v-for="i in 3" :key="i" class="sk-directory-row">
          <view class="sk-icon" />
          <view class="sk-col">
            <view class="sk-line sk-w45" />
            <view class="sk-line sk-w65" />
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'detail' | 'profile' | 'manage'
  }>(),
  { variant: 'detail' }
)
</script>

<style scoped>
.family-page-skeleton {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.sk-card {
  border: 1rpx solid rgba(255, 255, 255, 0.72);
  border-radius: var(--tree-radius-lg, 24rpx);
  background: rgba(255, 255, 255, 0.92);
  padding: 28rpx 30rpx;
  box-shadow: var(--tree-shadow-sm, 0 2rpx 12rpx rgba(15, 23, 42, 0.04));
}

.sk-card.sk-hero {
  padding-top: 34rpx;
  padding-bottom: 34rpx;
  background: rgba(31, 58, 95, 0.12);
  border-color: rgba(31, 58, 95, 0.08);
}

.sk-card.sk-soft {
  background: rgba(247, 250, 248, 0.94);
}

.sk-hero-head {
  display: flex;
  gap: 20rpx;
  margin-bottom: 20rpx;
}

.sk-circle {
  flex-shrink: 0;
  width: 88rpx;
  height: 88rpx;
  border-radius: 22rpx;
  background: rgba(148, 163, 184, 0.2);
}

.sk-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  min-width: 0;
}

.sk-line {
  height: 24rpx;
  border-radius: 8rpx;
  background: rgba(148, 163, 184, 0.22);
}

.sk-w25 { width: 25%; }
.sk-w30 { width: 30%; }
.sk-w35 { width: 35%; }
.sk-w40 { width: 40%; }
.sk-w45 { width: 45%; }
.sk-w50 { width: 50%; }
.sk-w55 { width: 55%; }
.sk-w60 { width: 60%; }
.sk-w65 { width: 65%; }
.sk-w70 { width: 70%; }
.sk-w90 { width: 90%; }

.sk-tags {
  display: flex;
  gap: 10rpx;
  margin-top: 4rpx;
}

.sk-pill {
  width: 96rpx;
  height: 32rpx;
  border-radius: 999rpx;
  background: rgba(148, 163, 184, 0.18);
}

.sk-pill--action {
  width: 120rpx;
  height: 44rpx;
}

.sk-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.1);
}

.sk-info-row:last-child {
  border-bottom: 0;
}

.sk-directory-row {
  display: flex;
  align-items: center;
  gap: 18rpx;
  min-height: 88rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.12);
}

.sk-directory-row:last-child {
  border-bottom: 0;
}

.sk-icon {
  flex-shrink: 0;
  width: 72rpx;
  height: 72rpx;
  border-radius: 20rpx;
  background: rgba(148, 163, 184, 0.16);
}

.sk-space-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 12rpx;
  padding-bottom: 16rpx;
  border-bottom: 1rpx solid rgba(148, 163, 184, 0.1);
}

.sk-context .sk-line {
  background: rgba(148, 163, 184, 0.18);
}
</style>
