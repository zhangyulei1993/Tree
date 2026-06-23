<template>
  <view class="tree-page legal-page">
    <MiniBackHome />
    <view class="legal-lead card">
      <text class="title">用户协议</text>
      <text class="muted">更新日期：2026-06-09 · 本页面内容将根据上线前法律审核结果更新。当前内容为产品说明占位，正式版本将以最终法律文本为准。</text>
    </view>
    <view v-for="section in sections" :key="section.title" class="card">
      <text class="section-title">{{ section.title }}</text>
      <text class="body">{{ section.content }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import MiniBackHome from '@/components/base/MiniBackHome.vue'

const sections = [
  { title: '一、服务说明', content: 'Tree 用于记录家庭成员、亲缘关系、公开家庭信息以及邀请和加入申请。用户应确保提交信息真实、合法，并获得必要授权。' },
  { title: '二、账号使用', content: '账号仅供本人使用。用户不得冒用他人身份，不得上传违法、有害或侵犯他人权益的内容。' },
  { title: '三、家庭数据', content: '平台账号与家庭成员节点相互独立。注销账号不会自动删除家庭树中的历史成员节点。' },
  { title: '四、公开展示', content: '家庭公开展示需经过申请和审核。公开内容可能被游客浏览，家庭管理员应谨慎选择公开信息。' },
  { title: '五、责任限制', content: '当前内容为产品说明占位，正式版本将以最终法律文本为准。如对协议内容有疑问，请在正式上线后以页面展示的最新版本为准。' }
]
</script>

<style scoped>
.legal-lead {
  margin-bottom: 20rpx;
}

.body {
  color: var(--tree-text-secondary);
  font-size: 27rpx;
  line-height: 1.9;
}
</style>
