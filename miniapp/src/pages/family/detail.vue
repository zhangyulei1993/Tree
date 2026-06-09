<template>
  <view class="page">
    <view class="card">
      <text class="title">{{ family.name }}</text>
      <text class="tag">FOUNDER</text>
      <text class="tag">{{ family.status }}</text>
      <view class="info-grid">
        <view><text class="label">姓氏</text><text>{{ family.surname }}</text></view>
        <view><text class="label">籍贯</text><text>{{ family.nativePlace }}</text></view>
        <view><text class="label">地区</text><text>{{ family.regionText }}</text></view>
        <view><text class="label">成员数</text><text>{{ family.memberCount }}</text></view>
      </view>
      <text class="muted">{{ family.description }}</text>
      <button class="button" @click="go('/pages/family/public-tree')">查看家庭树</button>
      <button class="button secondary" @click="go('/pages/invite/detail')">查看邀请示例</button>
    </view>
    <view class="card">
      <view class="section-row">
        <text class="section-title">成员概览</text>
        <text class="muted">graph v{{ family.graphVersion }}</text>
      </view>
      <view v-for="member in treeNodes" :key="member.memberId" class="member-row">
        <view>
          <text>{{ member.name }}</text>
          <text class="muted">{{ member.gender }} · {{ member.birthText }}</text>
        </view>
        <text class="tag">{{ member.memberId === 'member_001' ? 'FOUNDER' : 'MEMBER' }}</text>
      </view>
    </view>
    <view class="card">
      <text class="section-title">待处理事项</text>
      <view class="task-row">
        <text>加入申请</text>
        <text class="tag">1 条待处理</text>
      </view>
      <view class="task-row">
        <text>成员邀请</text>
        <text class="muted">暂无待处理</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { families, treeNodes } from '@/mock/data'

const family = families[0]

function go(url: string) {
  uni.navigateTo({ url })
}
</script>

<style scoped>
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin: 20rpx 0;
}

.info-grid view {
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.label {
  color: #6b7280;
  font-size: 22rpx;
}

.section-row,
.member-row,
.task-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.member-row,
.task-row {
  padding: 18rpx 0;
  border-top: 1rpx solid #e5e0d6;
}

.member-row view {
  display: flex;
  flex-direction: column;
}
</style>
