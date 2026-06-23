<template>
  <view class="tree-page legal-page">
    <MiniBackHome />
    <view class="legal-lead card">
      <text class="title">隐私政策</text>
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
  { title: '一、信息收集', content: '正式服务可能收集账号资料、手机号验证状态、家庭成员资料、邀请和申请记录，以及保障服务安全所需的设备与日志信息。' },
  { title: '二、信息使用', content: '相关信息用于账号认证、家庭关系维护、权限判断、公开展示审核、安全审计和用户主动发起的联系流程。' },
  { title: '三、敏感信息保护', content: '密码、验证码和登录凭证不以明文记录；微信登录会话标识、应用密钥及用户唯一标识不在页面或日志中完整展示。' },
  { title: '四、公开信息', content: '只有审核通过的家庭公开资料、公开树字段和已审核留言可被游客查看。手机号等账号信息不会作为公开树字段展示。' },
  { title: '五、用户权利', content: '用户可依法申请访问、更正或删除相关账号信息，也可发起账号注销。家庭历史节点的处理仍受家庭关系和审计规则约束。' }
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
